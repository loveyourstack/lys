package lys

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/loveyourstack/lys/lysauth"
)

type middlewareTestUser struct {
	id int64
}

func (u middlewareTestUser) GetUserId() int64 {
	return u.id
}

func TestLimitAuthed(t *testing.T) {
	middleware := LimitAuthed(lysauth.NewAppRateLimits(0, 1, time.Minute), discardLog)
	nextCalls := 0
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))

	newRequest := func(userID int64) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := WithUserInfo(req.Context(), middlewareTestUser{id: userID})
		return req.WithContext(ctx)
	}

	for _, tc := range []struct {
		userID int64
		status int
	}{
		{userID: 101, status: http.StatusNoContent},
		{userID: 101, status: http.StatusTooManyRequests},
		{userID: 202, status: http.StatusNoContent},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRequest(tc.userID))
		if w.Code != tc.status {
			t.Fatalf("user %d response status = %d, want %d", tc.userID, w.Code, tc.status)
		}
	}

	if nextCalls != 2 {
		t.Fatalf("next handler calls = %d, want 2", nextCalls)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(303))
	if w.Code != http.StatusNoContent {
		t.Fatalf("independent user response status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if nextCalls != 3 {
		t.Fatalf("next handler calls after independent user = %d, want 3", nextCalls)
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(101))
	if got := decodeStdResponse(t, w).ErrDescription; got != lysauth.ErrTooManyRequests.Message {
		t.Fatalf("limited response description = %q, want %q", got, lysauth.ErrTooManyRequests.Message)
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("missing user ID response status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if got := decodeStdResponse(t, w).ErrDescription; got != "An internal error occurred" {
		t.Fatalf("missing user ID response description = %q, want generic internal error", got)
	}
}

func TestLimitUnauthed(t *testing.T) {
	middleware := LimitUnauthed(lysauth.NewAppRateLimits(0, 1, time.Minute), discardLog)
	nextCalls := 0
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))
	ip := netip.MustParseAddr("203.0.113.1")
	newRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		return req.WithContext(lysauth.WithClientIp(req.Context(), ip))
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest())
	if w.Code != http.StatusNoContent {
		t.Fatalf("first request status = %d, want %d", w.Code, http.StatusNoContent)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest())
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("limited response status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
	if nextCalls != 1 {
		t.Fatalf("next handler calls = %d, want 1", nextCalls)
	}

	if got := decodeStdResponse(t, w).ErrDescription; got != lysauth.ErrTooManyRequests.Message {
		t.Fatalf("limited response description = %q, want %q", got, lysauth.ErrTooManyRequests.Message)
	}

	otherIPReq := httptest.NewRequest(http.MethodGet, "/", nil)
	otherIPReq = otherIPReq.WithContext(lysauth.WithClientIp(otherIPReq.Context(), netip.MustParseAddr("203.0.113.2")))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, otherIPReq)
	if w.Code != http.StatusNoContent {
		t.Fatalf("independent IP response status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if nextCalls != 2 {
		t.Fatalf("next handler calls after independent IP = %d, want 2", nextCalls)
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("missing client IP response status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if got := decodeStdResponse(t, w).ErrDescription; got != "An internal error occurred" {
		t.Fatalf("missing client IP response description = %q, want generic internal error", got)
	}
}

func TestRejectBlockedIp(t *testing.T) {
	loginAttempts := lysauth.NewAppLoginAttempts(3)
	blockedIP := netip.MustParseAddr("203.0.113.10")
	if err := loginAttempts.Block(blockedIP); err != nil {
		t.Fatalf("Block returned unexpected error: %v", err)
	}
	middleware := RejectBlockedIp(loginAttempts, discardLog)
	nextCalls := 0
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))
	newRequest := func(ip netip.Addr) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		return req.WithContext(lysauth.WithClientIp(req.Context(), ip))
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(netip.MustParseAddr("203.0.113.11")))
	if w.Code != http.StatusNoContent {
		t.Fatalf("unblocked response status = %d, want %d", w.Code, http.StatusNoContent)
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest(blockedIP))
	if w.Code != http.StatusForbidden {
		t.Fatalf("blocked response status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if got := decodeStdResponse(t, w).ErrDescription; got != lysauth.ErrBlocked.Message {
		t.Fatalf("blocked response description = %q, want %q", got, lysauth.ErrBlocked.Message)
	}
	if nextCalls != 1 {
		t.Fatalf("next handler calls = %d, want 1", nextCalls)
	}

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("missing client IP response status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestSetClientIp(t *testing.T) {
	for _, tc := range []struct {
		name            string
		useRealIpHeader bool
		remoteAddr      string
		realIP          string
		wantIP          netip.Addr
	}{
		{
			name:       "remote address",
			remoteAddr: "203.0.113.21:8080",
			wantIP:     netip.MustParseAddr("203.0.113.21"),
		},
		{
			name:            "real IP header",
			useRealIpHeader: true,
			remoteAddr:      "127.0.0.1:8080",
			realIP:          "203.0.113.20",
			wantIP:          netip.MustParseAddr("203.0.113.20"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := SetClientIp(tc.useRealIpHeader, discardLog)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotIP, err := lysauth.ClientIpFromCtx(r.Context())
				if err != nil {
					t.Errorf("ClientIpFromCtx returned unexpected error: %v", err)
				} else if gotIP != tc.wantIP {
					t.Errorf("client IP in context = %q, want %q", gotIP, tc.wantIP)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			req.Header.Set("X-Real-IP", tc.realIP)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusNoContent {
				t.Fatalf("valid request status = %d, want %d", w.Code, http.StatusNoContent)
			}
		})
	}

	handler := SetClientIp(true, discardLog)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("next handler called for invalid remote address")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "invalid"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("invalid remote address response status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if got := decodeStdResponse(t, w).ErrDescription; got != "An internal error occurred" {
		t.Fatalf("invalid remote address response description = %q, want generic internal error", got)
	}
}
