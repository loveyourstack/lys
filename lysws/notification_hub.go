package lysws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationHub manages WebSocket connections for user notifications.
// It listens for database notifications and broadcasts messages to connected clients based on user ID.
type NotificationHub struct {
	closed             atomic.Bool
	conns              map[int64][]*websocket.Conn // user_id → active sockets
	heartbeatPingIntvl time.Duration
	heartbeatPongWait  time.Duration
	heartbeatWriteWait time.Duration
	logger             *slog.Logger
	maxUserConnections int
	mu                 sync.RWMutex // protects conns
	upgrader           websocket.Upgrader

	// database fields for LISTEN/UNLISTEN
	db              *pgxpool.Pool
	dbListenConn    *pgxpool.Conn // single connection acquired from pool for LISTEN/UNLISTEN
	dbListenChannel string        // database channel to LISTEN on for notifications
}

type NotificationHubOptions struct {
	HeartbeatPingInterval time.Duration
	HeartbeatPongWait     time.Duration
	HeartbeatWriteWait    time.Duration
}

const (
	defaultHeartbeatPingIntvl = 54 * time.Second
	defaultHeartbeatPongWait  = 60 * time.Second
	defaultHeartbeatWriteWait = 10 * time.Second
)

// NewNotificationHub creates a new NotificationHub instance.
// It acquires a database connection for listening to notifications and initializes the connection map.
func NewNotificationHub(ctx context.Context, db *pgxpool.Pool, dbListenChannel string, maxUserConnections int,
	allowedOrigin string, logger *slog.Logger, options ...NotificationHubOptions) (hub *NotificationHub, err error) {

	if allowedOrigin == "" {
		return nil, fmt.Errorf("hub: allowedOrigin is required")
	}
	if db == nil {
		return nil, fmt.Errorf("hub: db is required")
	}
	if dbListenChannel == "" {
		return nil, fmt.Errorf("hub: dbListenChannel is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("hub: logger is required")
	}
	if maxUserConnections < 1 {
		return nil, fmt.Errorf("hub: maxUserConnections must be greater than 0")
	}

	opts := NotificationHubOptions{
		HeartbeatPingInterval: defaultHeartbeatPingIntvl,
		HeartbeatPongWait:     defaultHeartbeatPongWait,
		HeartbeatWriteWait:    defaultHeartbeatWriteWait,
	}
	if len(options) > 0 {
		opts = options[0]
		if opts.HeartbeatPingInterval == 0 {
			opts.HeartbeatPingInterval = defaultHeartbeatPingIntvl
		}
		if opts.HeartbeatPongWait == 0 {
			opts.HeartbeatPongWait = defaultHeartbeatPongWait
		}
		if opts.HeartbeatWriteWait == 0 {
			opts.HeartbeatWriteWait = defaultHeartbeatWriteWait
		}
	}
	if opts.HeartbeatPingInterval <= 0 {
		return nil, fmt.Errorf("heartbeatPingInterval must be greater than 0")
	}
	if opts.HeartbeatPongWait <= 0 {
		return nil, fmt.Errorf("heartbeatPongWait must be greater than 0")
	}
	if opts.HeartbeatWriteWait <= 0 {
		return nil, fmt.Errorf("heartbeatWriteWait must be greater than 0")
	}
	if opts.HeartbeatPingInterval >= opts.HeartbeatPongWait {
		return nil, fmt.Errorf("heartbeatPingInterval must be less than heartbeatPongWait")
	}

	// acquire a single connection from the pool for listening
	dbLisConn, err := db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("db.Acquire failed: %w", err)
	}

	// initialize the WebSocket upgrader with CORS check based on allowedOrigin
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == allowedOrigin
		},
	}

	return &NotificationHub{
		conns:              make(map[int64][]*websocket.Conn),
		heartbeatPingIntvl: opts.HeartbeatPingInterval,
		heartbeatPongWait:  opts.HeartbeatPongWait,
		heartbeatWriteWait: opts.HeartbeatWriteWait,
		logger:             logger.With("component", "notification hub"),
		maxUserConnections: maxUserConnections,
		upgrader:           upgrader,

		db:              db,
		dbListenConn:    dbLisConn,
		dbListenChannel: dbListenChannel,
	}, nil
}

func (h *NotificationHub) broadcast(userID int64, msg []byte, logFailures bool) (err error) {

	// copy user conns to avoid iteration issues due to Unregister calls
	h.mu.RLock()
	conns := h.conns[userID]
	connsCopy := make([]*websocket.Conn, len(conns))
	copy(connsCopy, conns)
	h.mu.RUnlock()

	for _, conn := range connsCopy {
		h.logger.Debug("broadcasting message", "user_id", userID, "message", string(msg))
		if connErr := conn.WriteMessage(websocket.TextMessage, msg); connErr != nil {
			if logFailures {
				h.logger.Error("conn.WriteMessage failed", "user_id", userID, "error", connErr)
			}
			h.Unregister(userID, conn)
			err = errors.Join(err, connErr)
		}
	}
	return err
}

// Broadcast sends a message to all active WebSocket connections for a given user ID.
func (h *NotificationHub) Broadcast(userID int64, msg []byte) {
	_ = h.broadcast(userID, msg, true)
}

// BroadcastE sends a message to all active WebSocket connections for a given user ID.
// It returns an error if any connection fails to send the message, but continues to attempt sending to all connections.
func (h *NotificationHub) BroadcastE(userID int64, msg []byte) (err error) {
	return h.broadcast(userID, msg, false)
}

// Close closes all active WebSocket connections and clears the connection map.
func (h *NotificationHub) Close() (err error) {

	// exit if already closed, and set to closed if not already
	if !h.closed.CompareAndSwap(false, true) {
		return nil
	}

	// snapshot all connections and clear conns map while holding the lock
	h.mu.Lock()
	all := make([]*websocket.Conn, 0)
	for _, conns := range h.conns {
		all = append(all, conns...)
	}
	h.conns = make(map[int64][]*websocket.Conn)
	h.mu.Unlock()

	// close the listen connection if it exists
	if h.dbListenConn != nil {
		h.dbListenConn.Release()
		h.dbListenConn = nil
	}

	// close sockets outside the lock so slow network closes do not block remaining ops
	for _, conn := range all {
		if closeErr := conn.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}

	return err
}

// GetUserConns returns a copy of the user's active WebSocket connections.
func (h *NotificationHub) GetUserConns(userID int64) []*websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return slices.Clone(h.conns[userID])
}

// NotificationSelectFunc defines a function type for selecting notification details from the database.
type NotificationSelectFunc func(ctx context.Context, db *pgxpool.Pool, notId int64) (userId int64, notType, message string, err error)

// ListenAndBroadcast listens for database notifications and broadcasts messages to users based on the notification payload.
// Only call this once per hub.
func (h *NotificationHub) ListenAndBroadcast(ctx context.Context, selectFunc NotificationSelectFunc) (err error) {

	if selectFunc == nil {
		return fmt.Errorf("selectFunc is required")
	}

	if h.closed.Load() {
		return fmt.Errorf("notification hub is closed")
	}

	// snapshot the listen connection during lock to avoid panic if Close is called while ListenAndBroadcast is running
	h.mu.Lock()
	lisConn := h.dbListenConn
	h.mu.Unlock()

	if lisConn == nil {
		return fmt.Errorf("dbListenConn is not initialized")
	}

	// LISTEN to receive notifications on the dbListenChannel
	_, err = lisConn.Exec(ctx, "LISTEN "+pgx.Identifier{h.dbListenChannel}.Sanitize())
	if err != nil {
		return fmt.Errorf("lisConn.Exec (LISTEN) failed on channel %s: %w", h.dbListenChannel, err)
	}
	defer func() {
		_, unlistenErr := lisConn.Exec(context.Background(), "UNLISTEN "+pgx.Identifier{h.dbListenChannel}.Sanitize())
		if unlistenErr != nil {
			h.logger.Error("lisConn.Exec (UNLISTEN) failed", "channel", h.dbListenChannel, "error", unlistenErr)
		}
	}()

	type messagePayload struct {
		Type string `json:"type"`
		Body string `json:"body"`
	}

	// wait for notifications or context cancellation
	for {
		not, err := lisConn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("lisConn.Conn().WaitForNotification failed: %w", err)
		}

		// payload needs to be the notification ID int64 to be looked up by selectFunc
		notId, err := strconv.ParseInt(not.Payload, 10, 64)
		if err != nil {
			h.logger.Error("strconv.ParseInt failed", "payload", not.Payload, "error", err)
			continue
		}

		// select the notification details
		userId, notType, message, err := selectFunc(ctx, h.db, notId)
		if err != nil {
			h.logger.Error("selectFunc failed", "notification_id", notId, "error", err)
			continue
		}

		// prepare JSON payload
		payload := messagePayload{
			Type: notType,
			Body: message,
		}
		msgBytes, err := json.Marshal(payload)
		if err != nil {
			h.logger.Error("json.Marshal failed", "payload", payload, "error", err)
			continue
		}

		// broadcast the message to the user's active connections
		if err := h.BroadcastE(userId, msgBytes); err != nil {
			h.logger.Error("h.BroadcastE failed", "user_id", userId, "error", err)
		}

	} // end for
}

// Register adds a WebSocket connection for a given user ID.
func (h *NotificationHub) Register(userID int64, c *websocket.Conn) (accepted bool) {

	if c == nil {
		h.logger.Error("connection cannot be nil")
		return false
	}

	var shouldClose bool
	var closeCode int
	var closeMsg string

	h.mu.Lock()

	switch {

	// reject if hub is closed
	case h.closed.Load():
		shouldClose = true
		h.logger.Error("notification hub is closed")

	// reject if c is already registered for userID (shouldn't happen but just in case)
	case slices.Contains(h.conns[userID], c):
		h.logger.Error("connection already registered for user", "user_id", userID)

	// reject if userID already has maximum active connections
	case len(h.conns[userID]) >= h.maxUserConnections:
		shouldClose = true

		// client should listen for this code to prevent endless reconnect loops
		closeCode = 4429
		closeMsg = "max connections reached"
		h.logger.Error("maximum active connections reached for user", "user_id", userID)

	// register new connection
	default:
		h.conns[userID] = append(h.conns[userID], c)
		h.logger.Debug("registered connection", "user_id", userID)
		accepted = true
	}

	h.mu.Unlock()

	if shouldClose {
		if closeCode != 0 || closeMsg != "" {
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, closeMsg))
			if err != nil {
				h.logger.Error("c.WriteMessage failed", "user_id", userID, "error", err)
			}
		}
		_ = c.Close()
	}

	return accepted
}

// ServeUserSocket upgrades the request to websocket, registers the user connection,
// and blocks until the connection closes while heartbeat ping/pong keeps dead peers short-lived.
func (h *NotificationHub) ServeUserSocket(ctx context.Context, w http.ResponseWriter, r *http.Request, userID int64) (err error) {

	// exit if hub is closed
	if h.closed.Load() {
		return fmt.Errorf("notification hub is closed")
	}

	// upgrade the HTTP request to a WebSocket connection
	wsConn, err := h.UpgradeHttpRequest(w, r)
	if err != nil {
		return fmt.Errorf("h.UpgradeHttpRequest failed: %w", err)
	}

	// register this connection for the user
	if !h.Register(userID, wsConn) {
		return nil
	}

	// ensure connection is unregistered when ServeUserSocket returns
	done := make(chan struct{})
	defer close(done)
	defer h.Unregister(userID, wsConn)

	// set initial read deadline for heartbeat
	if err := wsConn.SetReadDeadline(time.Now().Add(h.heartbeatPongWait)); err != nil {
		return fmt.Errorf("wsConn.SetReadDeadline failed: %w", err)
	}

	// set pong handler to update read deadline on pong receipt
	wsConn.SetPongHandler(func(string) error {
		return wsConn.SetReadDeadline(time.Now().Add(h.heartbeatPongWait))
	})

	// start a goroutine to send heartbeat pings at regular intervals
	go func() {
		ticker := time.NewTicker(h.heartbeatPingIntvl)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				// extend write deadline for ping message
				deadline := time.Now().Add(h.heartbeatWriteWait)

				// send ping and close connection if ping fails (e.g. due to network issues or client disconnect)
				if pingErr := wsConn.WriteControl(websocket.PingMessage, []byte("hb"), deadline); pingErr != nil {
					_ = wsConn.Close() // force read loop exit so deferred Unregister runs
					return
				}
			}
		}
	}()

	// block until connection is closed (e.g. by client, network issues, or user refreshing page)
	for {
		if _, _, err := wsConn.ReadMessage(); err != nil {
			return nil
		}
	}
}

// Status returns a snapshot of the number of active connections for each user ID.
func (h *NotificationHub) Status() map[int64]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status := make(map[int64]int, len(h.conns))
	for userID, conns := range h.conns {
		status[userID] = len(conns)
	}
	return status
}

// Unregister removes a WebSocket connection for a given user ID.
func (h *NotificationHub) Unregister(userID int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.conns[userID]
	for i, conn := range conns {
		if conn != c {
			continue
		}

		// close conn
		err := conn.Close()
		if err != nil {
			h.logger.Error("conn.Close failed", "user_id", userID, "error", err)
		}

		// remove conn from slice
		h.conns[userID] = append(conns[:i], conns[i+1:]...)

		// if no more connections for user, remove user from map
		if len(h.conns[userID]) == 0 {
			delete(h.conns, userID)
		}

		h.logger.Debug("unregistered connection", "user_id", userID)
		break
	}
}

// UpgradeHttpRequest upgrades an HTTP request to a WebSocket connection using the configured upgrader.
func (h *NotificationHub) UpgradeHttpRequest(w http.ResponseWriter, r *http.Request) (c *websocket.Conn, err error) {
	c, err = h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("h.upgrader.Upgrade failed: %w", err)
	}
	return c, nil
}

// UserConnCount returns the number of active connections for a given user ID.
func (h *NotificationHub) UserConnCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, exists := h.conns[userID]; exists {
		return len(conns)
	}
	return 0
}
