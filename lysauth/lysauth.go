package lysauth

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
)

// GetBearerToken returns the bearer token from the request's Authorization header.
func GetBearerToken(header http.Header) (token string, err error) {

	authHeaderVal := header.Get("Authorization")

	if authHeaderVal == "" {
		return "", lyserr.User{Message: "Authorization header missing or empty", StatusCode: http.StatusForbidden}
	}

	expectedPrefix := "Bearer "
	if !strings.HasPrefix(authHeaderVal, expectedPrefix) {
		return "", lyserr.User{Message: fmt.Sprintf("Authorization header value must start with '%s'", expectedPrefix), StatusCode: http.StatusForbidden}
	}

	expectedLen := len(expectedPrefix) + 1 // at least 1 character for token
	if len(authHeaderVal) < expectedLen {
		return "", lyserr.User{Message: fmt.Sprintf("Authorization header value too short, must be at least %d characters", expectedLen), StatusCode: http.StatusForbidden}
	}

	token = authHeaderVal[len(expectedPrefix):]

	return token, nil
}

// GetRemoteHostIP returns the remote IP from either the request RemoteAddr, or the X-Forwarded-For header.
// useXForwardedFor: set to true when using nginx reverse proxy or services like Cloudflare, since the remote IP will be in the X-Forwarded-For header instead of RemoteAddr.
// xForwardedForIdx: if using X-Forwarded-For header, this specifies which IP to use from the header (0 for first, etc).
// This is needed when there are multiple proxies and thus multiple IPs in the header.
func GetRemoteHostIP(r *http.Request, useXForwardedFor bool, xForwardedForIdx int) (remoteHostIP netip.Addr, err error) {

	if xForwardedForIdx < 0 {
		return netip.Addr{}, fmt.Errorf("invalid xForwardedForIdx: %d", xForwardedForIdx)
	}

	// get host ip from RemoteAddr
	ipStr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("net.SplitHostPort failed: %w", err)
	}

	// parse IP
	remoteHostIP, err = netip.ParseAddr(ipStr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid RemoteAddr IP: %s", ipStr)
	}

	// if using X-Forwarded-For, get host ip from that header instead
	if useXForwardedFor {

		// in lys projects, a reverse proxy is used in production, so RemoteAddr should be a loopback IP
		if !remoteHostIP.IsLoopback() {
			return netip.Addr{}, fmt.Errorf("remote IP %s is not loopback, but useXForwardedFor is true", remoteHostIP)
		}

		xForwardedVal := r.Header.Get("X-Forwarded-For")
		if xForwardedVal == "" {
			return netip.Addr{}, fmt.Errorf("X-Forwarded-For header is empty")
		}

		// assumes multiple X-Forwarded-For IPs are in the same header value, but split by comma

		// split header by comma
		xForwardedS := strings.Split(xForwardedVal, ",")
		if xForwardedForIdx > len(xForwardedS)-1 {
			return netip.Addr{}, fmt.Errorf("idx %d requested from X-Forwarded-For header, but len is only %d", xForwardedForIdx, len(xForwardedS))
		}

		// get IP from idx and trim whitespace
		remoteHostIPStr := strings.TrimSpace(xForwardedS[xForwardedForIdx])

		// parse X-Forwarded-For IP
		remoteHostIP, err = netip.ParseAddr(remoteHostIPStr)
		if err != nil {
			return netip.Addr{}, fmt.Errorf("invalid X-Forwarded-For IP: %s", remoteHostIPStr)
		}
	}

	// normalize IPv4-mapped IPv6 to match the canonical form stored in auth structs
	if remoteHostIP.Is4In6() {
		remoteHostIP = remoteHostIP.Unmap()
	}

	if !remoteHostIP.IsValid() {
		return netip.Addr{}, fmt.Errorf("empty remote IP")
	}

	return remoteHostIP, nil
}

// IsWebSocket returns true if the request Upgrade header contains "websocket".
func IsWebSocket(header http.Header) bool {

	upgradeHeaderVal := header.Get("Upgrade")
	if strings.EqualFold(upgradeHeaderVal, "websocket") {
		return true
	}

	return false
}
