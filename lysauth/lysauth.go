package lysauth

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
)

// GetBearerToken returns the bearer token from the provided header string. For HTTP, pass the value of the "Authorization" header.
func GetBearerToken(headerStr string) (token string, err error) {

	if headerStr == "" {
		return "", lyserr.User{Message: "Header missing or empty", StatusCode: http.StatusForbidden}
	}

	expectedPrefix := "Bearer "
	if !strings.HasPrefix(headerStr, expectedPrefix) {
		return "", lyserr.User{Message: fmt.Sprintf("Header value must start with '%s'", expectedPrefix), StatusCode: http.StatusForbidden}
	}

	expectedLen := len(expectedPrefix) + 1 // at least 1 character for token
	if len(headerStr) < expectedLen {
		return "", lyserr.User{Message: fmt.Sprintf("Header value too short - must be at least %d characters", expectedLen), StatusCode: http.StatusForbidden}
	}

	token = headerStr[len(expectedPrefix):]

	return token, nil
}

// GetIpAddrFromReq is an HTTP request wrapper for GetIpAddr.
func GetIpAddrFromReq(r *http.Request, useRealIpHeader bool) (remoteHostIP netip.Addr, err error) {
	return GetIpAddr(r.RemoteAddr, r.Header.Get("X-Real-IP"), useRealIpHeader)
}

// GetIpAddr returns the remote IP from either remoteAddr or xRealIpHeader.
// useRealIpHeader: set to true when using nginx reverse proxy actual client IP will be in the X-Real-IP header instead of RemoteAddr.
func GetIpAddr(remoteAddr, xRealIpHeader string, useRealIpHeader bool) (remoteHostIP netip.Addr, err error) {

	// get host ip from RemoteAddr
	ipStr, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("net.SplitHostPort failed: %w", err)
	}

	// parse IP
	remoteHostIP, err = netip.ParseAddr(ipStr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid RemoteAddr IP: %s", ipStr)
	}

	// if using X-Real-IP, get host ip from that header instead
	if useRealIpHeader {

		// in lys projects, a reverse proxy is used in production, so RemoteAddr should be a loopback IP
		if !remoteHostIP.IsLoopback() {
			return netip.Addr{}, fmt.Errorf("remote IP %s is not loopback, but useRealIpHeader is true", remoteHostIP)
		}

		xRealIpHeader = strings.TrimSpace(xRealIpHeader)
		if xRealIpHeader == "" {
			return netip.Addr{}, fmt.Errorf("X-Real-IP header is empty")
		}

		remoteHostIP, err = netip.ParseAddr(xRealIpHeader)
		if err != nil {
			return netip.Addr{}, fmt.Errorf("invalid X-Real-IP: %s", xRealIpHeader)
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
