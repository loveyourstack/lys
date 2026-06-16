package lysws

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testMaxUserConnections = 5

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestHub() *NotificationHub {
	return &NotificationHub{
		conns:              make(map[int64][]*websocket.Conn),
		errorLog:           testLogger(),
		heartbeatPingIntvl: defaultHeartbeatPingIntvl,
		heartbeatPongWait:  defaultHeartbeatPongWait,
		heartbeatWriteWait: defaultHeartbeatWriteWait,
		infoLog:            testLogger(),
		maxUserConnections: testMaxUserConnections,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func newWebsocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn, func()) {
	t.Helper()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	serverConnCh := make(chan *websocket.Conn, 1)
	errCh := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errCh <- err
			return
		}
		serverConnCh <- conn
	}))

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		ts.Close()
		t.Fatalf("dial websocket: %v", err)
	}

	var serverConn *websocket.Conn
	select {
	case serverConn = <-serverConnCh:
	case err = <-errCh:
		_ = clientConn.Close()
		ts.Close()
		t.Fatalf("upgrade websocket: %v", err)
	case <-time.After(2 * time.Second):
		_ = clientConn.Close()
		ts.Close()
		t.Fatal("timed out waiting for server websocket connection")
	}

	cleanup := func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
		ts.Close()
	}

	return serverConn, clientConn, cleanup
}

func TestBroadcastUnregistersFailedConnection(t *testing.T) {
	hub := newTestHub()
	userID := int64(42)

	goodServerConn, goodClientConn, cleanupGood := newWebsocketPair(t)
	t.Cleanup(cleanupGood)

	badServerConn, _, cleanupBad := newWebsocketPair(t)
	t.Cleanup(cleanupBad)

	if got := hub.Register(userID, goodServerConn); !got {
		t.Fatal("expected good connection registration to be accepted")
	}
	if got := hub.Register(userID, badServerConn); !got {
		t.Fatal("expected bad connection registration to be accepted")
	}

	if err := badServerConn.Close(); err != nil {
		t.Fatalf("close bad connection: %v", err)
	}

	hub.Broadcast(userID, []byte("hello"))

	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected failed connection to be unregistered; got %d connections", got)
	}

	if err := goodClientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	msgType, gotMsg, readErr := goodClientConn.ReadMessage()
	if readErr != nil {
		t.Fatalf("read broadcast message from good connection: %v", readErr)
	}
	if msgType != websocket.TextMessage {
		t.Fatalf("expected text message type, got %d", msgType)
	}
	if string(gotMsg) != "hello" {
		t.Fatalf("unexpected message body: got %q, want %q", string(gotMsg), "hello")
	}
}

func TestBroadcastEUnregistersFailedConnectionAndReturnsError(t *testing.T) {
	hub := newTestHub()
	userID := int64(42)

	goodServerConn, goodClientConn, cleanupGood := newWebsocketPair(t)
	t.Cleanup(cleanupGood)

	badServerConn, _, cleanupBad := newWebsocketPair(t)
	t.Cleanup(cleanupBad)

	if got := hub.Register(userID, goodServerConn); !got {
		t.Fatal("expected good connection registration to be accepted")
	}
	if got := hub.Register(userID, badServerConn); !got {
		t.Fatal("expected bad connection registration to be accepted")
	}

	if err := badServerConn.Close(); err != nil {
		t.Fatalf("close bad connection: %v", err)
	}

	msg := []byte("hello")
	err := hub.BroadcastE(userID, msg)
	if err == nil {
		t.Fatal("expected broadcast error when one connection is closed")
	}

	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected failed connection to be unregistered; got %d connections", got)
	}

	if err := goodClientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	msgType, gotMsg, readErr := goodClientConn.ReadMessage()
	if readErr != nil {
		t.Fatalf("read broadcast message from good connection: %v", readErr)
	}
	if msgType != websocket.TextMessage {
		t.Fatalf("expected text message type, got %d", msgType)
	}
	if string(gotMsg) != string(msg) {
		t.Fatalf("unexpected message body: got %q, want %q", string(gotMsg), string(msg))
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	hub := newTestHub()
	userID := int64(11)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if got := hub.Register(userID, serverConn); !got {
		t.Fatal("expected connection registration to be accepted")
	}

	if err := hub.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := hub.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}

	if !hub.closed.Load() {
		t.Fatal("expected hub to remain marked closed after second close")
	}
}

func TestCloseMarksHubClosedAndClearsConnections(t *testing.T) {
	hub := newTestHub()
	userID := int64(77)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if got := hub.Register(userID, serverConn); !got {
		t.Fatal("expected connection registration to be accepted")
	}

	if err := hub.Close(); err != nil {
		t.Fatalf("close hub: %v", err)
	}

	if !hub.closed.Load() {
		t.Fatal("expected hub to be marked closed")
	}

	hub.mu.RLock()
	connCount := len(hub.conns)
	hub.mu.RUnlock()
	if connCount != 0 {
		t.Fatalf("expected connection map to be cleared; got %d user entries", connCount)
	}
}

func TestGetUserConnsReturnsCopyOfConnections(t *testing.T) {
	hub := newTestHub()
	userID := int64(200)

	if got := hub.GetUserConns(userID); len(got) != 0 {
		t.Fatalf("expected empty slice for unknown user, got %d", len(got))
	}

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if got := hub.Register(userID, serverConn); !got {
		t.Fatal("expected connection registration to be accepted")
	}

	conns := hub.GetUserConns(userID)
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0] != serverConn {
		t.Fatal("expected returned connection to match registered connection")
	}

	// Mutating the returned slice must not affect hub internals.
	conns[0] = nil
	if hub.UserConnCount(userID) != 1 {
		t.Fatal("mutating returned slice should not affect hub connection count")
	}
}

func TestListenAndBroadcastValidation(t *testing.T) {
	dummySelect := func(ctx context.Context, db *pgxpool.Pool, notID int64) (int64, string, string, error) {
		return 1, "x", "y", nil
	}

	t.Run("nil select func", func(t *testing.T) {
		hub := newTestHub()
		err := hub.ListenAndBroadcast(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "selectFunc is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("closed hub", func(t *testing.T) {
		hub := newTestHub()
		hub.closed.Store(true)

		err := hub.ListenAndBroadcast(context.Background(), dummySelect)
		if err == nil || !strings.Contains(err.Error(), "notification hub is closed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing listen conn", func(t *testing.T) {
		hub := newTestHub()
		hub.dbListenChannel = "chan_notifications"

		err := hub.ListenAndBroadcast(context.Background(), dummySelect)
		if err == nil || !strings.Contains(err.Error(), "dbListenConn is not initialized") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewNotificationHubValidation(t *testing.T) {
	tests := []struct {
		name               string
		db                 *pgxpool.Pool
		allowedOrigin      string
		channel            string
		maxUserConnections int
		infoLog            *slog.Logger
		errorLog           *slog.Logger
		wantErr            string
	}{
		{
			name:               "empty allowedOrigin",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "",
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "allowedOrigin is required",
		},
		{
			name:               "nil db",
			db:                 nil,
			allowedOrigin:      "*",
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "db is required",
		},
		{
			name:               "empty channel",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "*",
			channel:            "",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "dbListenChannel is required",
		},
		{
			name:               "maxUserConnections zero",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "*",
			channel:            "chan_notifications",
			maxUserConnections: 0,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "maxUserConnections must be greater than 0",
		},
		{
			name:               "maxUserConnections negative",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "*",
			channel:            "chan_notifications",
			maxUserConnections: -1,
			infoLog:            testLogger(),
			errorLog:           testLogger(),
			wantErr:            "maxUserConnections must be greater than 0",
		},
		{
			name:               "nil infoLog",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "*",
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            nil,
			errorLog:           testLogger(),
			wantErr:            "infoLog is required",
		},
		{
			name:               "nil errorLog",
			db:                 &pgxpool.Pool{},
			allowedOrigin:      "*",
			channel:            "chan_notifications",
			maxUserConnections: testMaxUserConnections,
			infoLog:            testLogger(),
			errorLog:           nil,
			wantErr:            "errorLog is required",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hub, err := NewNotificationHub(context.Background(), tc.db, tc.channel, tc.maxUserConnections, tc.allowedOrigin, tc.infoLog, tc.errorLog)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if hub != nil {
				t.Fatal("expected nil hub on constructor validation failure")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: got %q, want contains %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestRegisterAndUnregister(t *testing.T) {
	hub := newTestHub()
	userID := int64(99)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if got := hub.Register(userID, serverConn); !got {
		t.Fatal("expected connection registration to be accepted")
	}
	if got := hub.UserConnCount(userID); got != 1 {
		t.Fatalf("expected 1 registered connection, got %d", got)
	}

	hub.Unregister(userID, serverConn)
	if got := hub.UserConnCount(userID); got != 0 {
		t.Fatalf("expected 0 connections after unregister, got %d", got)
	}

	hub.mu.RLock()
	_, exists := hub.conns[userID]
	hub.mu.RUnlock()
	if exists {
		t.Fatal("expected user entry to be removed after last connection unregister")
	}
}

func TestRegisterEnforcesMaxUserConnections(t *testing.T) {
	hub := newTestHub()
	userID := int64(123)

	cleanups := make([]func(), 0, testMaxUserConnections+1)
	t.Cleanup(func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	})

	for i := 0; i < testMaxUserConnections; i++ {
		serverConn, _, cleanup := newWebsocketPair(t)
		cleanups = append(cleanups, cleanup)

		if got := hub.Register(userID, serverConn); !got {
			t.Fatalf("expected registration %d to be accepted", i)
		}
	}

	extraConn, extraClientConn, cleanup := newWebsocketPair(t)
	cleanups = append(cleanups, cleanup)

	if got := hub.Register(userID, extraConn); got {
		t.Fatal("expected registration to be rejected when max active connections is reached")
	}

	if got := hub.UserConnCount(userID); got != testMaxUserConnections {
		t.Fatalf("expected %d connections after max enforcement, got %d", testMaxUserConnections, got)
	}

	if err := extraClientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, _, readErr := extraClientConn.ReadMessage()
	if readErr == nil {
		t.Fatal("expected close error from extra rejected connection")
	}

	closeErr, ok := readErr.(*websocket.CloseError)
	if !ok {
		t.Fatalf("expected websocket close error, got %T (%v)", readErr, readErr)
	}

	if closeErr.Code != 4429 {
		t.Fatalf("expected close code 4429, got %d", closeErr.Code)
	}
	if closeErr.Text != "max connections reached" {
		t.Fatalf("expected close text %q, got %q", "max connections reached", closeErr.Text)
	}
}

func TestRegisterNilConnectionIsIgnored(t *testing.T) {
	hub := newTestHub()

	if got := hub.Register(1, nil); got {
		t.Fatal("expected nil connection registration to be rejected")
	}

	if got := hub.UserConnCount(1); got != 0 {
		t.Fatalf("expected nil connection to be ignored, got %d", got)
	}
}

func TestRegisterRejectsWhenHubClosed(t *testing.T) {
	hub := newTestHub()
	hub.closed.Store(true)

	serverConn, _, cleanup := newWebsocketPair(t)
	t.Cleanup(cleanup)

	if got := hub.Register(7, serverConn); got {
		t.Fatal("expected registration to be rejected when hub is closed")
	}

	if got := hub.UserConnCount(7); got != 0 {
		t.Fatalf("expected no registration when hub is closed, got %d", got)
	}

	if writeErr := serverConn.WriteMessage(websocket.TextMessage, []byte("x")); writeErr == nil {
		t.Fatal("expected rejected connection to be closed when hub is closed")
	}
}

func TestServeUserSocketRegistersAndUnregistersOnClientDisconnect(t *testing.T) {
	hub := newTestHub()
	userID := int64(500)

	serverErrCh := make(chan error, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrCh <- hub.ServeUserSocket(r.Context(), w, r, userID)
	}))
	t.Cleanup(ts.Close)

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })

	deadline := time.Now().Add(2 * time.Second)
	for hub.UserConnCount(userID) != 1 {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for registration, got %d", hub.UserConnCount(userID))
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatalf("close client websocket: %v", err)
	}

	deadline = time.Now().Add(2 * time.Second)
	for hub.UserConnCount(userID) != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for unregister, got %d", hub.UserConnCount(userID))
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case serveErr := <-serverErrCh:
		if serveErr != nil {
			t.Fatalf("ServeUserSocket returned unexpected error: %v", serveErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ServeUserSocket to exit")
	}
}

func TestServeUserSocketRejectsWhenHubClosed(t *testing.T) {
	hub := newTestHub()
	hub.closed.Store(true)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := hub.ServeUserSocket(r.Context(), w, r, 1)
		if err == nil || !strings.Contains(err.Error(), "notification hub is closed") {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	t.Cleanup(ts.Close)

	// A plain HTTP GET is enough to trigger the handler; no WebSocket upgrade needed.
	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	_ = resp.Body.Close()
}

func TestStatusReturnsConnectionCountsPerUser(t *testing.T) {
	hub := newTestHub()

	if got := hub.Status(); len(got) != 0 {
		t.Fatalf("expected empty status for empty hub, got %v", got)
	}

	userA := int64(301)
	userB := int64(302)

	connA1, _, cleanupA1 := newWebsocketPair(t)
	t.Cleanup(cleanupA1)
	connA2, _, cleanupA2 := newWebsocketPair(t)
	t.Cleanup(cleanupA2)
	connB1, _, cleanupB1 := newWebsocketPair(t)
	t.Cleanup(cleanupB1)

	hub.Register(userA, connA1)
	hub.Register(userA, connA2)
	hub.Register(userB, connB1)

	status := hub.Status()
	if status[userA] != 2 {
		t.Fatalf("expected 2 connections for userA, got %d", status[userA])
	}
	if status[userB] != 1 {
		t.Fatalf("expected 1 connection for userB, got %d", status[userB])
	}

	// Mutating the returned map must not affect hub internals.
	delete(status, userA)
	if hub.UserConnCount(userA) != 2 {
		t.Fatal("mutating returned status map should not affect hub connection count")
	}
}
