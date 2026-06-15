package lysauth

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"slices"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lysstring"
	"github.com/loveyourstack/lys/lystype"
)

// SessionInput is a user session input.
type SessionInput struct {
	AllowMultipleSessions bool       `json:"allow_multiple_sessions"`
	FamilyName            string     `json:"family_name" validate:"required"`
	ForcePasswordChange   bool       `json:"force_password_change"`
	GeoIpCountryIsoCode   string     `json:"geo_ip_country_iso_code" validate:"required,len=2"`
	GeoIpLocation         string     `json:"geo_ip_location" validate:"required"`
	GivenName             string     `json:"given_name" validate:"required"`
	Ip                    netip.Addr `json:"ip" validate:"required"`
	Roles                 []string   `json:"roles" validate:"required"`
	UserAgent             string     `json:"user_agent" validate:"required"`
	UserId                int64      `json:"user_id" validate:"required,gt=0"`
	UserName              string     `json:"user_name" validate:"required"`
}

// Session is a user session.
type Session struct {
	CreatedAt    lystype.Datetime `json:"created_at"`
	ExpiresAt    lystype.Datetime `json:"expires_at"`
	LastAccessAt lystype.Datetime `json:"last_access_at"`
	Token        string           `json:"-"`
	SessionInput
}

// AppSessions contains sessions and methods to manage them.
type AppSessions struct {
	all              map[string]Session // map of token to session
	mu               sync.RWMutex
	sessionDuration  time.Duration // duration of a session before it expires
	useXForwardedFor bool          // whether to use X-Forwarded-For header to determine IP
	validate         *validator.Validate
	xForwardedForIdx int // if using X-Forwarded-For, which index to use (0 for first, etc)
}

// NewAppSessions creates a new AppSessions instance.
func NewAppSessions(validate *validator.Validate, sessionDuration time.Duration, useXForwardedFor bool, xForwardedForIdx int) *AppSessions {
	return &AppSessions{
		all: make(map[string]Session),

		sessionDuration:  sessionDuration,
		useXForwardedFor: useXForwardedFor,
		validate:         validate,
		xForwardedForIdx: xForwardedForIdx,
	}
}

// Add validates and creates a new session in appSessions.
func (appS *AppSessions) Add(sessionInput SessionInput) (token string, err error) {

	// validate params
	err = lysmeta.Validate(appS.validate, sessionInput)
	if err != nil {
		return "", fmt.Errorf("lysmeta.Validate failed: %w", err)
	}
	if !sessionInput.Ip.IsValid() {
		return "", fmt.Errorf("empty IP")
	}

	// normalize ipv4-mapped IPv6 addresses to IPv4
	if sessionInput.Ip.Is4In6() {
		sessionInput.Ip = sessionInput.Ip.Unmap()
	}

	// if user doesn't allow multiple sessions, delete existing sessions for this user
	if !sessionInput.AllowMultipleSessions {
		appS.DeleteByUserId(sessionInput.UserId)
	}

	// add session
	token = lysstring.Rand(24)
	sess := Session{
		SessionInput: sessionInput,

		CreatedAt:    lystype.Datetime(time.Now()),
		ExpiresAt:    lystype.Datetime(time.Now().Add(appS.sessionDuration)),
		LastAccessAt: lystype.Datetime(time.Now()),
		Token:        token,
	}
	appS.mu.Lock()
	appS.all[token] = sess
	appS.mu.Unlock()

	return token, nil
}

// All returns all sessions.
func (appS *AppSessions) All() (sessions []Session) {
	appS.mu.RLock()
	defer appS.mu.RUnlock()

	for _, session := range appS.all {
		sessions = append(sessions, session)
	}
	return sessions
}

// Count returns the number of active sessions.
func (appS *AppSessions) Count() int {
	appS.mu.RLock()
	defer appS.mu.RUnlock()
	return len(appS.all)
}

// DeleteByIp deletes all sessions for the specified IP address.
func (appS *AppSessions) DeleteByIp(ip netip.Addr) error {
	if !ip.IsValid() {
		return fmt.Errorf("empty IP")
	}

	// normalize IPv4-mapped IPv6 to match the canonical form stored by Add.
	if ip.Is4In6() {
		ip = ip.Unmap()
	}

	appS.mu.Lock()
	defer appS.mu.Unlock()

	for token, session := range appS.all {
		if session.Ip == ip {
			delete(appS.all, token)
		}
	}

	return nil
}

// DeleteByTokens deletes the sessions for the specified tokens.
func (appS *AppSessions) DeleteByTokens(tokens []string) {
	appS.mu.Lock()
	defer appS.mu.Unlock()

	for _, token := range tokens {
		_, exists := appS.all[token]
		if !exists {
			continue
		}
		delete(appS.all, token)
	}
}

// DeleteByUserId deletes all sessions for the specified user ID.
func (appS *AppSessions) DeleteByUserId(userId int64) {
	appS.mu.Lock()
	defer appS.mu.Unlock()

	for token, session := range appS.all {
		if session.UserId == userId {
			delete(appS.all, token)
		}
	}
}

// FromRequest returns the session associated with the request, or an error if the session is invalid.
func (appS *AppSessions) FromRequest(r *http.Request, infoLog *slog.Logger) (sess Session, err error) {

	isWebSocket := IsWebSocket(r.Header)
	token := ""

	if isWebSocket {
		// ws: get token from query param
		token = r.URL.Query().Get("token")
		if token == "" {
			return Session{}, fmt.Errorf("ws: token param is empty or missing")
		}
	} else {
		// http: get token from req auth header
		token, err = GetBearerToken(r.Header)
		if err != nil {
			return Session{}, fmt.Errorf("GetBearerToken failed: %w", err)
		}
	}

	// get IP
	remoteHostIp, err := GetRemoteHostIP(r, appS.useXForwardedFor, appS.xForwardedForIdx)
	if err != nil {
		return Session{}, fmt.Errorf("GetRemoteHostIP failed: %w", err)
	}

	// find token in app sessions
	appS.mu.RLock()
	session, exists := appS.all[token]
	appS.mu.RUnlock()
	if !exists {
		return Session{}, lyserr.User{Message: "token not found", StatusCode: http.StatusForbidden}
	}

	// make sure request IP matches session IP
	if remoteHostIp != session.Ip {
		infoLog.Debug(fmt.Sprintf("session IP mismatch: remoteHostIp=%s, sessionIp=%s", remoteHostIp, session.Ip))
		return Session{}, lyserr.User{Message: "invalid token", StatusCode: http.StatusForbidden}
	}

	// http only: make sure request UserAgent matches session UserAgent
	if !isWebSocket && r.UserAgent() != session.UserAgent {
		infoLog.Debug(fmt.Sprintf("session UserAgent mismatch: requestUserAgent=%s, sessionUserAgent=%s", r.UserAgent(), session.UserAgent))
		return Session{}, lyserr.User{Message: "invalid token", StatusCode: http.StatusForbidden}
	}

	// check if session has expired
	if time.Now().After(time.Time(session.ExpiresAt)) {
		return Session{}, lyserr.User{Message: "session expired", StatusCode: http.StatusForbidden}
	}

	// session verified, http only: update LastAccessAt and ExpiresAt
	if !isWebSocket {
		session.LastAccessAt = lystype.Datetime(time.Now())
		session.ExpiresAt = lystype.Datetime(time.Now().Add(appS.sessionDuration))
		appS.mu.Lock()
		appS.all[token] = session
		appS.mu.Unlock()
	}

	return session, nil
}

// GetByUserId returns all sessions for the specified user ID.
func (appS *AppSessions) GetByUserId(userId int64) (sessions []Session) {
	appS.mu.RLock()
	defer appS.mu.RUnlock()

	for _, session := range appS.all {
		if session.UserId == userId {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetExpired returns all expired sessions.
func (appS *AppSessions) GetExpired() (sessions []Session) {
	appS.mu.RLock()
	defer appS.mu.RUnlock()

	now := time.Now()
	for _, session := range appS.all {
		if now.After(time.Time(session.ExpiresAt)) {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// ListByLastAccessAt returns all sessions sorted by LastAccessAt.
func (appS *AppSessions) ListByLastAccessAt(asc bool) (sortedSessions []Session) {

	appS.mu.RLock()
	defer appS.mu.RUnlock()

	if len(appS.all) == 0 {
		return []Session{} // for JSON encoding to return [] instead of null
	}

	for _, session := range appS.all {
		sortedSessions = append(sortedSessions, session)
	}

	slices.SortFunc(sortedSessions, func(a, b Session) int {
		if time.Time(a.LastAccessAt).Equal(time.Time(b.LastAccessAt)) {
			return 0
		}
		if time.Time(a.LastAccessAt).Before(time.Time(b.LastAccessAt)) {
			return -1
		}
		return 1
	})
	if asc {
		return sortedSessions
	}

	slices.Reverse(sortedSessions)
	return sortedSessions
}

// Load loads the supplied sessions into a freshly created AppSessions instance.
func (appS *AppSessions) Load(sessions []Session) (err error) {
	if len(sessions) == 0 {
		return fmt.Errorf("sessions has len 0")
	}

	appS.mu.Lock()
	defer appS.mu.Unlock()

	if appS.all == nil {
		return fmt.Errorf("AppSessions is not initialized")
	}
	if len(appS.all) > 0 {
		return fmt.Errorf("expected empty AppSessions instance, but instance has %d sessions", len(appS.all))
	}

	for _, session := range sessions {
		appS.all[session.Token] = session
	}

	return nil
}
