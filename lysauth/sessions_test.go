package lysauth

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lystype"
)

func newAuthRequest(t *testing.T, ip, token, userAgent string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.RemoteAddr = net.JoinHostPort(ip, "12345")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", userAgent)

	return req
}

func newWebSocketAuthRequest(t *testing.T, ip, token, userAgent string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://example.com/ws?token="+token, nil)
	if err != nil {
		t.Fatalf("failed creating websocket request: %v", err)
	}
	req.RemoteAddr = net.JoinHostPort(ip, "12345")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("User-Agent", userAgent)

	return req
}

func newDefaultSessionInput() SessionInput {
	return SessionInput{
		AllowMultipleSessions: false,
		FamilyName:            "Doe",
		ForcePasswordChange:   false,
		GeoIpCountryIsoCode:   "US",
		GeoIpLocation:         "New York, NY",
		GivenName:             "Jane",
		Ip:                    netip.MustParseAddr("198.51.100.100"),
		Roles:                 []string{"Tech"},
		UserAgent:             "test-agent",
		UserId:                1,
		UserName:              "jane.doe",
	}
}

func TestAppSessions_Add_RejectsZeroValueIP(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	input := newDefaultSessionInput()
	input.Ip = netip.Addr{}
	input.UserName = "zero-ip"

	_, err := appS.Add(input)
	if err == nil {
		t.Fatalf("Add expected empty IP error, got nil")
	}
	if !strings.Contains(err.Error(), "empty IP") {
		t.Fatalf("Add error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}
}

func TestAppSessions_Add_NormalizesMappedIPv4(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	input := newDefaultSessionInput()
	input.Ip = netip.MustParseAddr("::ffff:198.51.100.77")
	input.UserAgent = "ua-normalize"
	input.UserId = 77
	input.UserName = "normalize"

	token, err := appS.Add(input)
	if err != nil {
		t.Fatalf("Add returned unexpected error: %v", err)
	}

	appS.mu.RLock()
	sess := appS.all[token]
	appS.mu.RUnlock()

	if sess.Ip != netip.MustParseAddr("198.51.100.77") {
		t.Fatalf("stored session IP mismatch: got %q, want %q", sess.Ip, netip.MustParseAddr("198.51.100.77"))
	}
	if sess.Ip.Is4In6() {
		t.Fatalf("stored session IP should be unmapped IPv4, got mapped IPv6 %q", sess.Ip)
	}
}

func TestAppSessions_Add(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	if got := appS.Count(); got != 0 {
		t.Fatalf("initial Count mismatch: got %d, want 0", got)
	}

	validInput := newDefaultSessionInput()
	validInput.Ip = netip.MustParseAddr("203.0.113.12")
	validInput.UserId = 12
	validInput.UserName = "james"

	token, err := appS.Add(validInput)
	if err != nil {
		t.Fatalf("Add(valid) returned unexpected error: %v", err)
	}
	if token == "" {
		t.Fatalf("Add(valid) token mismatch: got empty token")
	}
	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after Add(valid): got %d, want 1", got)
	}

	tests := []struct {
		name    string
		input   SessionInput
		wantErr string
	}{
		{
			name: "missing user name",
			input: func() SessionInput {
				i := newDefaultSessionInput()
				i.Ip = netip.MustParseAddr("203.0.113.20")
				i.UserName = ""
				return i
			}(),
			wantErr: "lysmeta.Validate failed",
		},
		{
			name: "invalid user id",
			input: func() SessionInput {
				i := newDefaultSessionInput()
				i.Ip = netip.MustParseAddr("203.0.113.21")
				i.UserId = 0
				i.UserName = "user"
				return i
			}(),
			wantErr: "lysmeta.Validate failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := appS.Add(tc.input)
			if err == nil {
				t.Fatalf("Add expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Add error mismatch: got %q, want contains %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestAppSessions_Add_SingleSessionPolicyDeletesPriorSessions(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	firstInput := newDefaultSessionInput()
	firstInput.AllowMultipleSessions = false
	firstInput.Ip = netip.MustParseAddr("203.0.113.60")
	firstInput.UserAgent = "agent-first"
	firstInput.UserId = 55
	firstInput.UserName = "single-user"

	firstToken, err := appS.Add(firstInput)
	if err != nil {
		t.Fatalf("Add(first) returned unexpected error: %v", err)
	}
	if firstToken == "" {
		t.Fatalf("Add(first) token mismatch: got empty token")
	}

	secondInput := newDefaultSessionInput()
	secondInput.AllowMultipleSessions = false
	secondInput.Ip = netip.MustParseAddr("203.0.113.61")
	secondInput.UserAgent = "agent-second"
	secondInput.UserId = 55
	secondInput.UserName = "single-user"

	secondToken, err := appS.Add(secondInput)
	if err != nil {
		t.Fatalf("Add(second) returned unexpected error: %v", err)
	}
	if secondToken == "" {
		t.Fatalf("Add(second) token mismatch: got empty token")
	}
	if secondToken == firstToken {
		t.Fatalf("Add(second) token mismatch: got same token as first session")
	}

	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after replacing existing session: got %d, want 1", got)
	}

	reqOld, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed creating old-token request: %v", err)
	}
	reqOld.RemoteAddr = firstInput.Ip.String() + ":12345"
	reqOld.Header.Set("Authorization", "Bearer "+firstToken)
	reqOld.Header.Set("User-Agent", firstInput.UserAgent)

	_, err = appS.FromRequest(reqOld, slog.Default())
	if err == nil {
		t.Fatalf("FromRequest(old token) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token not found") {
		t.Fatalf("FromRequest(old token) error mismatch: got %q, want contains %q", err.Error(), "token not found")
	}

	reqNew, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed creating new-token request: %v", err)
	}
	reqNew.RemoteAddr = secondInput.Ip.String() + ":12345"
	reqNew.Header.Set("Authorization", "Bearer "+secondToken)
	reqNew.Header.Set("User-Agent", secondInput.UserAgent)

	if _, err = appS.FromRequest(reqNew, slog.Default()); err != nil {
		t.Fatalf("FromRequest(new token) returned unexpected error: %v", err)
	}
}

func TestAppSessions_DeleteByUserId(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	input1 := newDefaultSessionInput()
	input1.AllowMultipleSessions = true
	input1.Ip = netip.MustParseAddr("198.51.100.10")
	input1.UserAgent = "ua-1"
	input1.UserId = 10
	input1.UserName = "user-10"

	_, err := appS.Add(input1)
	if err != nil {
		t.Fatalf("setup Add(user-10 first) failed: %v", err)
	}
	input2 := newDefaultSessionInput()
	input2.AllowMultipleSessions = true
	input2.Ip = netip.MustParseAddr("198.51.100.11")
	input2.UserAgent = "ua-2"
	input2.UserId = 10
	input2.UserName = "user-10"

	_, err = appS.Add(input2)
	if err != nil {
		t.Fatalf("setup Add(user-10 second) failed: %v", err)
	}
	input3 := newDefaultSessionInput()
	input3.Ip = netip.MustParseAddr("198.51.100.12")
	input3.UserAgent = "ua-3"
	input3.UserId = 11
	input3.UserName = "user-11"

	_, err = appS.Add(input3)
	if err != nil {
		t.Fatalf("setup Add(user-11) failed: %v", err)
	}

	appS.DeleteByUserId(10)

	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after DeleteByUserId: got %d, want 1", got)
	}
}

func TestAppSessions_DeleteByToken(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	input1 := newDefaultSessionInput()
	input1.Ip = netip.MustParseAddr("198.51.100.30")
	input1.UserAgent = "ua-30"
	input1.UserId = 30
	input1.UserName = "user-30"

	token1, err := appS.Add(input1)
	if err != nil {
		t.Fatalf("setup Add(user-30) failed: %v", err)
	}
	input2 := newDefaultSessionInput()
	input2.AllowMultipleSessions = true
	input2.Ip = netip.MustParseAddr("198.51.100.31")
	input2.UserAgent = "ua-31"
	input2.UserId = 31
	input2.UserName = "user-31"

	token2, err := appS.Add(input2)
	if err != nil {
		t.Fatalf("setup Add(user-31) failed: %v", err)
	}

	if got := appS.Count(); got != 2 {
		t.Fatalf("Count mismatch after setup: got %d, want 2", got)
	}

	appS.DeleteByTokens([]string{token1})
	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after DeleteByToken(existing): got %d, want 1", got)
	}

	_, err = appS.FromRequest(newAuthRequest(t, "198.51.100.30", token1, "ua-30"), slog.Default())
	if err == nil {
		t.Fatalf("FromRequest(deleted token) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token not found") {
		t.Fatalf("FromRequest(deleted token) error mismatch: got %q, want contains %q", err.Error(), "token not found")
	}

	appS.DeleteByTokens([]string{"missing-token"})

	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after DeleteByToken(missing): got %d, want 1", got)
	}

	if _, err = appS.FromRequest(newAuthRequest(t, "198.51.100.31", token2, "ua-31"), slog.Default()); err != nil {
		t.Fatalf("FromRequest(remaining token) returned unexpected error: %v", err)
	}
}

func TestAppSessions_DeleteByIp(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	input1 := newDefaultSessionInput()
	input1.AllowMultipleSessions = true
	input1.Ip = netip.MustParseAddr("198.51.100.30")
	input1.UserAgent = "ua-del-1"
	input1.UserId = 130
	input1.UserName = "user-130"

	_, err := appS.Add(input1)
	if err != nil {
		t.Fatalf("setup Add(user-130) failed: %v", err)
	}
	input2 := newDefaultSessionInput()
	input2.AllowMultipleSessions = true
	input2.Ip = netip.MustParseAddr("198.51.100.31")
	input2.UserAgent = "ua-del-2"
	input2.UserId = 131
	input2.UserName = "user-131"

	_, err = appS.Add(input2)
	if err != nil {
		t.Fatalf("setup Add(user-131) failed: %v", err)
	}

	if got := appS.Count(); got != 2 {
		t.Fatalf("Count mismatch after setup: got %d, want 2", got)
	}

	err = appS.DeleteByIp(netip.MustParseAddr("::ffff:198.51.100.30"))
	if err != nil {
		t.Fatalf("DeleteByIp(mapped IPv4) returned unexpected error: %v", err)
	}
	if got := appS.Count(); got != 1 {
		t.Fatalf("Count mismatch after DeleteByIp(mapped IPv4): got %d, want 1", got)
	}

	err = appS.DeleteByIp(netip.Addr{})
	if err == nil {
		t.Fatalf("DeleteByIp(zero-value IP) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "empty IP") {
		t.Fatalf("DeleteByIp(zero-value IP) error mismatch: got %q, want contains %q", err.Error(), "empty IP")
	}
}

func TestAppSessions_FromRequest(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	sessionInput := newDefaultSessionInput()
	sessionInput.Ip = netip.MustParseAddr("198.51.100.44")
	sessionInput.UserAgent = "agent-1"
	sessionInput.UserId = 101
	sessionInput.UserName = "alice"
	token, err := appS.Add(sessionInput)
	if err != nil {
		t.Fatalf("setup Add returned unexpected error: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.RemoteAddr = sessionInput.Ip.String() + ":12345"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", sessionInput.UserAgent)

	sess, err := appS.FromRequest(req, slog.Default())
	if err != nil {
		t.Fatalf("FromRequest(valid) returned unexpected error: %v", err)
	}
	if sess.UserId != sessionInput.UserId {
		t.Fatalf("FromRequest(valid) UserId mismatch: got %d, want %d", sess.UserId, sessionInput.UserId)
	}
	if sess.UserName != sessionInput.UserName {
		t.Fatalf("FromRequest(valid) UserName mismatch: got %q, want %q", sess.UserName, sessionInput.UserName)
	}
	if sess.Ip != sessionInput.Ip {
		t.Fatalf("FromRequest(valid) Ip mismatch: got %q, want %q", sess.Ip, sessionInput.Ip)
	}
}

func TestAppSessions_FromRequest_AcceptsMappedIPv4(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	sessionInput := newDefaultSessionInput()
	sessionInput.Ip = netip.MustParseAddr("198.51.100.44")
	sessionInput.UserAgent = "agent-mapped"
	sessionInput.UserId = 201
	sessionInput.UserName = "mapped-user"
	token, err := appS.Add(sessionInput)
	if err != nil {
		t.Fatalf("setup Add returned unexpected error: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.RemoteAddr = "[::ffff:198.51.100.44]:12345"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", sessionInput.UserAgent)

	if _, err = appS.FromRequest(req, slog.Default()); err != nil {
		t.Fatalf("FromRequest(mapped ipv4 remote) returned unexpected error: %v", err)
	}
}

func TestAppSessions_FromRequest_WebSocket(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	sessionInput := newDefaultSessionInput()
	sessionInput.Ip = netip.MustParseAddr("198.51.100.120")
	sessionInput.UserAgent = "ua-http"
	sessionInput.UserId = 6120
	sessionInput.UserName = "ws-user"

	token, err := appS.Add(sessionInput)
	if err != nil {
		t.Fatalf("setup Add returned unexpected error: %v", err)
	}

	t.Run("valid websocket request uses token query param", func(t *testing.T) {
		req := newWebSocketAuthRequest(t, sessionInput.Ip.String(), token, "different-ua")

		sess, err := appS.FromRequest(req, slog.Default())
		if err != nil {
			t.Fatalf("FromRequest(valid websocket) returned unexpected error: %v", err)
		}
		if sess.UserId != sessionInput.UserId {
			t.Fatalf("FromRequest(valid websocket) UserId mismatch: got %d, want %d", sess.UserId, sessionInput.UserId)
		}
	})

	t.Run("websocket ignores Authorization header", func(t *testing.T) {
		req := newWebSocketAuthRequest(t, sessionInput.Ip.String(), token, sessionInput.UserAgent)
		req.Header.Set("Authorization", "Bearer wrong-token")

		if _, err = appS.FromRequest(req, slog.Default()); err != nil {
			t.Fatalf("FromRequest(websocket with wrong Authorization header) returned unexpected error: %v", err)
		}
	})

	t.Run("missing token query param", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com/ws", nil)
		if err != nil {
			t.Fatalf("failed creating websocket request: %v", err)
		}
		req.RemoteAddr = net.JoinHostPort(sessionInput.Ip.String(), "12345")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("User-Agent", sessionInput.UserAgent)

		_, err = appS.FromRequest(req, slog.Default())
		if err == nil {
			t.Fatalf("FromRequest(websocket missing token query) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "ws: token param is empty or missing") {
			t.Fatalf("FromRequest(websocket missing token query) error mismatch: got %q, want contains %q", err.Error(), "ws: token param is empty or missing")
		}
	})
}

func TestAppSessions_FromRequest_XForwardedFor(t *testing.T) {
	validate := validator.New()

	t.Run("success with selected forwarded ip", func(t *testing.T) {
		appS := NewAppSessions(validate, 10*time.Hour, true, 1)
		input := newDefaultSessionInput()
		input.Ip = netip.MustParseAddr("198.51.100.90")
		input.UserAgent = "xff-ok"
		input.UserId = 901
		input.UserName = "xff-ok"

		token, err := appS.Add(input)
		if err != nil {
			t.Fatalf("setup Add returned unexpected error: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.RemoteAddr = "127.0.0.1:443"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.90")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", input.UserAgent)

		if _, err = appS.FromRequest(req, slog.Default()); err != nil {
			t.Fatalf("FromRequest(X-Forwarded-For success) returned unexpected error: %v", err)
		}
	})

	t.Run("remote must be loopback when using forwarded header", func(t *testing.T) {
		appS := NewAppSessions(validate, 10*time.Hour, true, 0)
		input := newDefaultSessionInput()
		input.Ip = netip.MustParseAddr("198.51.100.91")
		input.UserAgent = "xff-loopback"
		input.UserId = 902
		input.UserName = "xff-loopback"

		token, err := appS.Add(input)
		if err != nil {
			t.Fatalf("setup Add returned unexpected error: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.RemoteAddr = "10.0.0.7:443"
		req.Header.Set("X-Forwarded-For", "198.51.100.91")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", input.UserAgent)

		_, err = appS.FromRequest(req, slog.Default())
		if err == nil {
			t.Fatalf("FromRequest(non-loopback reverse proxy) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "GetRemoteHostIP failed") {
			t.Fatalf("FromRequest(non-loopback reverse proxy) error mismatch: got %q, want contains %q", err.Error(), "GetRemoteHostIP failed")
		}
		if !strings.Contains(err.Error(), "not loopback") {
			t.Fatalf("FromRequest(non-loopback reverse proxy) error mismatch: got %q, want contains %q", err.Error(), "not loopback")
		}
	})

	t.Run("forwarded index out of range", func(t *testing.T) {
		appS := NewAppSessions(validate, 10*time.Hour, true, 2)
		input := newDefaultSessionInput()
		input.Ip = netip.MustParseAddr("198.51.100.92")
		input.UserAgent = "xff-idx"
		input.UserId = 903
		input.UserName = "xff-idx"

		token, err := appS.Add(input)
		if err != nil {
			t.Fatalf("setup Add returned unexpected error: %v", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("failed creating request: %v", err)
		}
		req.RemoteAddr = "127.0.0.1:443"
		req.Header.Set("X-Forwarded-For", "198.51.100.92")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", input.UserAgent)

		_, err = appS.FromRequest(req, slog.Default())
		if err == nil {
			t.Fatalf("FromRequest(X-Forwarded-For idx out of range) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "GetRemoteHostIP failed") {
			t.Fatalf("FromRequest(X-Forwarded-For idx out of range) error mismatch: got %q, want contains %q", err.Error(), "GetRemoteHostIP failed")
		}
		if !strings.Contains(err.Error(), "idx 2 requested") {
			t.Fatalf("FromRequest(X-Forwarded-For idx out of range) error mismatch: got %q, want contains %q", err.Error(), "idx 2 requested")
		}
	})
}

func TestAppSessions_FromRequestErrors(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	sessionInput := newDefaultSessionInput()
	sessionInput.Ip = netip.MustParseAddr("203.0.113.50")
	sessionInput.Roles = []string{"Senior"}
	sessionInput.UserAgent = "agent-2"
	sessionInput.UserId = 300
	sessionInput.UserName = "bob"
	token, err := appS.Add(sessionInput)
	if err != nil {
		t.Fatalf("setup Add returned unexpected error: %v", err)
	}

	tests := []struct {
		name        string
		buildReq    func() *http.Request
		mutate      func()
		wantErr     string
		wantUserErr bool
	}{
		{
			name: "missing auth header",
			buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.RemoteAddr = sessionInput.Ip.String() + ":9999"
				req.Header.Set("User-Agent", sessionInput.UserAgent)
				return req
			},
			wantErr: "GetBearerToken failed",
		},
		{
			name: "token not found",
			buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.RemoteAddr = sessionInput.Ip.String() + ":9999"
				req.Header.Set("Authorization", "Bearer not-found-token")
				req.Header.Set("User-Agent", sessionInput.UserAgent)
				return req
			},
			wantErr:     "token not found",
			wantUserErr: true,
		},
		{
			name: "different ip",
			buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.RemoteAddr = "203.0.113.51:9999"
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("User-Agent", sessionInput.UserAgent)
				return req
			},
			wantErr:     "invalid token",
			wantUserErr: true,
		},
		{
			name: "different user agent",
			buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.RemoteAddr = sessionInput.Ip.String() + ":9999"
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("User-Agent", "other-agent")
				return req
			},
			wantErr:     "invalid token",
			wantUserErr: true,
		},
		{
			name: "session expired",
			buildReq: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.RemoteAddr = sessionInput.Ip.String() + ":9999"
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("User-Agent", sessionInput.UserAgent)
				return req
			},
			mutate: func() {
				appS.mu.Lock()
				s := appS.all[token]
				s.ExpiresAt = lystype.Datetime(time.Now().Add(-1 * time.Minute))
				appS.all[token] = s
				appS.mu.Unlock()
			},
			wantErr:     "session expired",
			wantUserErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mutate != nil {
				tc.mutate()
			}

			req := tc.buildReq()
			_, err := appS.FromRequest(req, slog.Default())
			if err == nil {
				t.Fatalf("FromRequest expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("FromRequest error mismatch: got %q, want contains %q", err.Error(), tc.wantErr)
			}

			if tc.wantUserErr {
				var userErr lyserr.User
				if !errors.As(err, &userErr) {
					t.Fatalf("expected lyserr.User, got %T", err)
				}
			}
		})
	}
}

func TestAppSessions_ListByLastAccessAt(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	if got := appS.ListByLastAccessAt(true); len(got) != 0 {
		t.Fatalf("ListByLastAccessAt on empty map mismatch: got non-empty slice, want empty")
	}

	t1 := time.Now().Add(-3 * time.Hour)
	t2 := time.Now().Add(-2 * time.Hour)
	t3 := time.Now().Add(-1 * time.Hour)

	appS.mu.Lock()
	appS.all["tok-1"] = Session{
		LastAccessAt: lystype.Datetime(t2),
		SessionInput: func() SessionInput {
			i := newDefaultSessionInput()
			i.Ip = netip.MustParseAddr("198.51.100.1")
			i.UserAgent = "ua1"
			i.UserId = 1
			i.UserName = "u1"
			return i
		}()}
	appS.all["tok-2"] = Session{
		LastAccessAt: lystype.Datetime(t1),
		SessionInput: func() SessionInput {
			i := newDefaultSessionInput()
			i.Ip = netip.MustParseAddr("198.51.100.2")
			i.UserAgent = "ua2"
			i.UserId = 2
			i.UserName = "u2"
			return i
		}()}
	appS.all["tok-3"] = Session{
		LastAccessAt: lystype.Datetime(t3),
		SessionInput: func() SessionInput {
			i := newDefaultSessionInput()
			i.Ip = netip.MustParseAddr("198.51.100.3")
			i.UserAgent = "ua3"
			i.UserId = 3
			i.UserName = "u3"
			return i
		}()}
	appS.mu.Unlock()

	asc := appS.ListByLastAccessAt(true)
	if len(asc) != 3 {
		t.Fatalf("ListByLastAccessAt asc len mismatch: got %d, want 3", len(asc))
	}
	if time.Time(asc[0].LastAccessAt) != t1 || time.Time(asc[1].LastAccessAt) != t2 || time.Time(asc[2].LastAccessAt) != t3 {
		t.Fatalf("ListByLastAccessAt asc order mismatch")
	}

	desc := appS.ListByLastAccessAt(false)
	if len(desc) != 3 {
		t.Fatalf("ListByLastAccessAt desc len mismatch: got %d, want 3", len(desc))
	}
	if time.Time(desc[0].LastAccessAt) != t3 || time.Time(desc[1].LastAccessAt) != t2 || time.Time(desc[2].LastAccessAt) != t1 {
		t.Fatalf("ListByLastAccessAt desc order mismatch")
	}
}

func TestAppSessions_ConcurrentAddAndFromRequest(t *testing.T) {
	validate := validator.New()
	appS := NewAppSessions(validate, 10*time.Hour, false, 0)

	type sessionRef struct {
		token     string
		ip        string
		userAgent string
	}

	const totalAdds = 120
	refs := make(chan sessionRef, totalAdds)
	errCh := make(chan error, totalAdds*2)

	var addWG sync.WaitGroup
	for i := range totalAdds {
		addWG.Go(func() {

			ip := fmt.Sprintf("203.0.113.%d", (i%200)+1)
			ua := fmt.Sprintf("agent-%d", i)

			input := newDefaultSessionInput()
			input.Ip = netip.MustParseAddr(ip)
			input.UserAgent = ua
			input.UserId = int64(i + 1)
			input.UserName = fmt.Sprintf("user-%d", i)

			token, err := appS.Add(input)
			if err != nil {
				errCh <- fmt.Errorf("Add(%d) failed: %w", i, err)
				return
			}

			refs <- sessionRef{token: token, ip: ip, userAgent: ua}
		})
	}

	var readWG sync.WaitGroup
	const readers = 8
	for range readers {
		readWG.Go(func() {

			for ref := range refs {
				req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
				if err != nil {
					errCh <- fmt.Errorf("NewRequest failed: %w", err)
					continue
				}

				req.RemoteAddr = ref.ip + ":12345"
				req.Header.Set("Authorization", "Bearer "+ref.token)
				req.Header.Set("User-Agent", ref.userAgent)

				if _, err = appS.FromRequest(req, slog.Default()); err != nil {
					errCh <- fmt.Errorf("FromRequest failed: %w", err)
				}
			}
		})
	}

	addWG.Wait()
	close(refs)
	readWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}

	if got := appS.Count(); got != totalAdds {
		t.Fatalf("Count mismatch after concurrent Add/FromRequest: got %d, want %d", got, totalAdds)
	}
}
