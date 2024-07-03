package lys

import (
	"context"
	"net/http"

	"github.com/loveyourstack/lys/lysstring"
)

type contextKey string

// ReqUserInfoCtxKey is the key to be used when binding a ReqUserInfo to a request via context
const ReqUserInfoCtxKey contextKey = "ReqUserInfoKey"

// ReqUserInfo contains data about the API user which is added to request context after authentication
type ReqUserInfo struct {
	Roles    []string `json:"roles"`
	UserId   int64    `json:"user_id"`
	UserName string   `json:"user_name"`
}

// GetUserFromCtx returns the user from ctx. ReqUserInfoCtxKey and ReqUserInfo must have been assigned in middleware
func GetUserFromCtx(ctx context.Context) ReqUserInfo {

	userInfo, ok := ctx.Value(ReqUserInfoCtxKey).(ReqUserInfo)
	if !ok {
		return ReqUserInfo{}
	}
	return userInfo
}

// GetUserNameFromCtx returns the username if it can be obtained from ctx, otherwise the supplied default value
func GetUserNameFromCtx(ctx context.Context, defaultVal string) string {

	userInfo, ok := ctx.Value(ReqUserInfoCtxKey).(ReqUserInfo)
	if !ok {
		return defaultVal
	}
	return userInfo.UserName
}

// AuthorizeRole is middleware that checks that the user has one of the supplied allowedRoles
// is intended for use in subroutes
func AuthorizeRole(allowedRoles []string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// get the user info from context
			userInfo, ok := r.Context().Value(ReqUserInfoCtxKey).(ReqUserInfo)
			if !ok {
				HandleUserError(http.StatusForbidden, ErrDescUserInfoMissing, w)
				return
			}

			// check user is authorized to do this
			if len(allowedRoles) > 0 && !lysstring.ContainsAny(userInfo.Roles, allowedRoles) {
				HandleUserError(http.StatusForbidden, ErrDescPermissionDenied, w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
