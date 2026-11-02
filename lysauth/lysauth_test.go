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
			wantErr:  "Header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header has no values",
			headers: http.Header{
				"Authorization": []string{},
			},
			wantErr:  "Header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header value empty",
			headers: http.Header{
				"Authorization": []string{""},
			},
			wantErr:  "Header missing or empty",
			wantUser: true,
		},
		{
			name: "authorization header wrong prefix",
			headers: http.Header{
				"Authorization": []string{"Token abc123"},
			},
			wantErr:  "Header value must start with 'Bearer '",
			wantUser: true,
		},
		{
			name: "authorization header too short",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantErr:  "Header value too short - must be at least 8 characters",
			wantUser: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tc.headers.Get("Authorization"))
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
				if _, ok := errors.AsType[lyserr.User](err); !ok {
					t.Fatalf("expected lyserr.User, got %T", err)
				}
			}
		})
	}
}

func TestGetIpAddr(t *testing.T) {
	tests := []struct {
		name            string
		remoteAddr      string
		xRealIpHeader   string
		useRealIpHeader bool
		wantIP          netip.Addr
		wantErr         string
	}{
		{
			name:       "remote addr success without real ip header",
			remoteAddr: "10.2.3.4:8080",
			wantIP:     netip.MustParseAddr("10.2.3.4"),
		},
		{
			name:       "remote ipv6 success without real ip header",
			remoteAddr: "[2001:db8::11]:8080",
			wantIP:     netip.MustParseAddr("2001:db8::11"),
		},
		{
			name:       "remote ipv4-mapped ipv6 is normalized",
			remoteAddr: "[::ffff:198.51.100.42]:8080",
			wantIP:     netip.MustParseAddr("198.51.100.42"),
		},
		{
			name:       "remote addr missing port",
			remoteAddr: "10.2.3.4",
			wantErr:    "net.SplitHostPort failed",
		},
		{
			name:       "remote addr invalid ip",
			remoteAddr: "not-an-ip:8080",
			wantErr:    "invalid RemoteAddr IP: not-an-ip",
		},
		{
			name:            "real ip header requires loopback proxy",
			remoteAddr:      "10.2.3.4:8080",
			xRealIpHeader:   "203.0.113.8",
			useRealIpHeader: true,
			wantErr:         "remote IP 10.2.3.4 is not loopback, but useRealIpHeader is true",
		},
		{
			name:            "real ip header missing",
			remoteAddr:      "127.0.0.1:8080",
			useRealIpHeader: true,
			wantErr:         "X-Real-IP header is empty",
		},
		{
			name:            "real ip header invalid",
			remoteAddr:      "127.0.0.1:8080",
			xRealIpHeader:   "bad-ip",
			useRealIpHeader: true,
			wantErr:         "invalid X-Real-IP: bad-ip",
		},
		{
			name:            "real ip header is trimmed",
			remoteAddr:      "127.0.0.1:8080",
			xRealIpHeader:   " 203.0.113.7 ",
			useRealIpHeader: true,
			wantIP:          netip.MustParseAddr("203.0.113.7"),
		},
		{
			name:            "real ip header ipv6",
			remoteAddr:      "[::1]:8080",
			xRealIpHeader:   "2001:db8::20",
			useRealIpHeader: true,
			wantIP:          netip.MustParseAddr("2001:db8::20"),
		},
		{
			name:            "real ip header mapped ipv4 is normalized",
			remoteAddr:      "127.0.0.1:8080",
			xRealIpHeader:   "::ffff:198.51.100.44",
			useRealIpHeader: true,
			wantIP:          netip.MustParseAddr("198.51.100.44"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIP, err := GetIpAddr(tc.remoteAddr, tc.xRealIpHeader, tc.useRealIpHeader)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("GetIpAddr returned unexpected error: %v", err)
				}
				if gotIP != tc.wantIP {
					t.Fatalf("GetIpAddr IP mismatch: got %q, want %q", gotIP, tc.wantIP)
				}
				return
			}

			if err == nil {
				t.Fatalf("GetIpAddr expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("GetIpAddr error mismatch: got %q, want contains %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestGetIpAddrFromReq(t *testing.T) {
	tests := []struct {
		name            string
		remoteAddr      string
		xRealIpHeader   string
		useRealIpHeader bool
		wantIP          netip.Addr
	}{
		{
			name:          "uses remote addr when real ip header disabled",
			remoteAddr:    "198.51.100.1:8080",
			xRealIpHeader: "203.0.113.1",
			wantIP:        netip.MustParseAddr("198.51.100.1"),
		},
		{
			name:            "uses real ip header when enabled",
			remoteAddr:      "127.0.0.1:8080",
			xRealIpHeader:   "203.0.113.1",
			useRealIpHeader: true,
			wantIP:          netip.MustParseAddr("203.0.113.1"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if err != nil {
				t.Fatalf("failed creating request: %v", err)
			}
			req.RemoteAddr = tc.remoteAddr
			req.Header.Set("X-Real-IP", tc.xRealIpHeader)

			gotIP, err := GetIpAddrFromReq(req, tc.useRealIpHeader)
			if err != nil {
				t.Fatalf("GetIpAddrFromReq returned unexpected error: %v", err)
			}
			if gotIP != tc.wantIP {
				t.Fatalf("GetIpAddrFromReq IP mismatch: got %q, want %q", gotIP, tc.wantIP)
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
