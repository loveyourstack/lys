package lys

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/loveyourstack/lys/lysauth"
)

// LimitAuthed is middleware that applies rate limiting to authenticated users based on user ID.
func LimitAuthed(rateLimits *lysauth.AppRateLimits, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// get user info from context if a GetUserId() method is available
			userId := GetUserIdFromCtx(ctx)
			if userId == 0 {
				HandleInternalError(ctx, fmt.Errorf("LimitAuthed: user ID missing"), logger, w)
				return
			}
			key := strconv.FormatInt(userId, 10)

			// check rate limit
			if !rateLimits.Allow(key) {
				HandleUserError(lysauth.ErrTooManyRequests, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LimitUnauthed is middleware that applies rate limiting to unauthenticated users based on IP.
func LimitUnauthed(rateLimits *lysauth.AppRateLimits, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ip, err := lysauth.ClientIpFromCtx(ctx)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("LimitUnauthed: %w", err), logger, w)
				return
			}

			// check rate limit
			if !rateLimits.Allow(ip.String()) {
				HandleUserError(lysauth.ErrTooManyRequests, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RejectBlockedIp is middleware that rejects requests from blocked IPs based on login attempts.
func RejectBlockedIp(loginAttempts *lysauth.AppLoginAttempts, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			remoteHostIP, err := lysauth.ClientIpFromCtx(ctx)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("RejectBlockedIp: %w", err), logger, w)
				return
			}

			isBlocked, err := loginAttempts.IsBlocked(remoteHostIP)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("RejectBlockedIp: loginAttempts.IsBlocked failed for IP %s: %w", remoteHostIP, err), logger, w)
				return
			}

			// reject if blocked
			if isBlocked {
				HandleUserError(lysauth.ErrBlocked, w)
				logger.Debug(fmt.Sprintf("RejectBlockedIp: blocked request from IP %s - %s %s %s", remoteHostIP, r.Proto, r.Method, r.URL.EscapedPath()))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetClientIp returns middleware that resolves the client IP and stores it in context.
func SetClientIp(useRealIpHeader bool, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			remoteHostIp, err := lysauth.GetIpAddrFromReq(r, useRealIpHeader)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("SetClientIp: lysauth.GetIpAddrFromReq failed: %w", err), logger, w)
				return
			}
			ctx = lysauth.WithClientIp(ctx, remoteHostIp)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
