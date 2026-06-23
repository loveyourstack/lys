package lysauth

import (
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"testing"

	"github.com/loveyourstack/lys/lyserr"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   string
		wantUser  bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123"},
			},
			wantToken: "abc123",
		},
		{
			name:     "missing authorization header",
			headers:  http.Header{},
			wantErr:  "Authorization header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header has no values",
			headers: http.Header{
				"Authorization": []string{},
			},
			wantErr:  "Authorization header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header value empty",
			headers: http.Header{
				"Authorization": []string{""},
			},
			wantErr:  "Authorization header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header wrong prefix",
			headers: http.Header{
				"Authorization": []string{"Token abc123"},
			},
			wantErr:  "Authorization header value must start with 'Bearer '",
			wantUser: true,
		},
		{
			name: "authorization header too short",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantErr:  "Authorization header value too short, must be at least 8 characters",
			wantUser: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tc.headers)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("GetBearerToken returned unexpected error: %v", err)
				}
				if gotToken != tc.wantToken {
					t.Fatalf("GetBearerToken token mismatch: got %q, want %q", gotToken, tc.wantToken)
				}
				return
			}

			if err == nil {
				t.Fatalf("GetBearerToken expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("GetBearerToken error mismatch: got %q, want %q", err.Error(), tc.wantErr)
			}

			if tc.wantUser {
				var userErr lyserr.User
				if !errors.As(err, &userErr) {
					t.Fatalf("expected lyserr.User, got %T", err)
				}
			}
		})
	}
}

func TestGetRemoteHostIP(t *testing.T) {
	tests := []struct {
		name             string
		remoteAddr       string
		headers          http.Header
		useXForwardedFor bool
		xForwardedForIdx int
		wantIP           netip.Addr
		wantErr          string
	}{
		{
			name:             "negative x-forwarded-for idx",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: -1,
			wantErr:          "invalid xForwardedForIdx: -1",
		},
		{
			name:             "remote addr success without x-forwarded-for",
			remoteAddr:       "10.2.3.4:8080",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: 0,
			wantIP:           netip.MustParseAddr("10.2.3.4"),
		},
		{
			name:             "remote ipv6 success without x-forwarded-for",
			remoteAddr:       "[2001:db8::11]:8080",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: 0,
			wantIP:           netip.MustParseAddr("2001:db8::11"),
		},
		{
			name:             "remote ipv4-mapped ipv6 is normalized",
			remoteAddr:       "[::ffff:198.51.100.42]:8080",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: 0,
			wantIP:           netip.MustParseAddr("198.51.100.42"),
		},
		{
			name:             "remote addr missing port",
			remoteAddr:       "10.2.3.4",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: 0,
			wantErr:          "net.SplitHostPort failed",
		},
		{
			name:             "remote addr invalid ip",
			remoteAddr:       "not-an-ip:8080",
			headers:          http.Header{},
			useXForwardedFor: false,
			xForwardedForIdx: 0,
			wantErr:          "invalid RemoteAddr IP: not-an-ip",
		},
		{
			name:             "x-forwarded-for requires loopback proxy",
			remoteAddr:       "10.2.3.4:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"203.0.113.8"}},
			useXForwardedFor: true,
			xForwardedForIdx: 0,
			wantErr:          "remote IP 10.2.3.4 is not loopback, but useXForwardedFor is true",
		},
		{
			name:             "x-forwarded-for missing header",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{},
			useXForwardedFor: true,
			xForwardedForIdx: 0,
			wantErr:          "X-Forwarded-For header is empty",
		},
		{
			name:             "x-forwarded-for idx out of range",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"198.51.100.1"}},
			useXForwardedFor: true,
			xForwardedForIdx: 1,
			wantErr:          "idx 1 requested from X-Forwarded-For header, but len is only 1",
		},
		{
			name:             "x-forwarded-for selected ip invalid",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"bad-ip"}},
			useXForwardedFor: true,
			xForwardedForIdx: 0,
			wantErr:          "invalid X-Forwarded-For IP: bad-ip",
		},
		{
			name:             "x-forwarded-for selected ip with spaces",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"203.0.113.7, 198.51.100.5"}},
			useXForwardedFor: true,
			xForwardedForIdx: 1,
			wantIP:           netip.MustParseAddr("198.51.100.5"),
		},
		{
			name:             "x-forwarded-for selected ipv6",
			remoteAddr:       "[::1]:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"2001:db8::10, 2001:db8::20"}},
			useXForwardedFor: true,
			xForwardedForIdx: 1,
			wantIP:           netip.MustParseAddr("2001:db8::20"),
		},
		{
			name:             "x-forwarded-for selected mapped ipv4 is normalized",
			remoteAddr:       "127.0.0.1:8080",
			headers:          http.Header{"X-Forwarded-For": []string{"::ffff:198.51.100.44"}},
			useXForwardedFor: true,
			xForwardedForIdx: 0,
			wantIP:           netip.MustParseAddr("198.51.100.44"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.RemoteAddr = tc.remoteAddr
			req.Header = tc.headers

			gotIP, err := GetRemoteHostIP(req, tc.useXForwardedFor, tc.xForwardedForIdx)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("GetRemoteHostIP returned unexpected error: %v", err)
				}
				if gotIP != tc.wantIP {
					t.Fatalf("GetRemoteHostIP IP mismatch: got %q, want %q", gotIP, tc.wantIP)
				}
				return
			}

			if err == nil {
				t.Fatalf("GetRemoteHostIP expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("GetRemoteHostIP error mismatch: got %q, want contains %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestIsWebSocket(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    bool
	}{
		{
			name: "upgrade websocket",
			headers: http.Header{
				"Upgrade": []string{"websocket"},
			},
			want: true,
		},
		{
			name:    "no upgrade header",
			headers: http.Header{},
			want:    false,
		},
		{
			name: "upgrade header no values",
			headers: http.Header{
				"Upgrade": []string{},
			},
			want: false,
		},
		{
			name: "upgrade value case insensitive",
			headers: http.Header{
				"Upgrade": []string{"WebSocket"},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsWebSocket(tc.headers)
			if got != tc.want {
				t.Fatalf("IsWebSocket mismatch: got %v, want %v", got, tc.want)
			}
		})
	}
}
