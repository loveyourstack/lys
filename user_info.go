package lys

import (
	"context"

	"github.com/loveyourstack/lys/lyspgdb"
)

// UserInfoCtxKey is the key that should be used when binding a user info struct to a request via context
// if you use this key and add a GetUserName() string method to the struct, the username will be included in error logs when using the error handlers in error_handlers.go.
// if you add a GetUserId() int64 method to the struct, GetUserIdFromCtx will be available to use in your code.
const UserInfoCtxKey lyspgdb.ContextKey = "UserInfoKey"

/*

a sample user info struct and methods:

type ReqUserInfo struct {
	Roles    []string `json:"roles"`
	UserId   int64    `json:"user_id"`
	UserName string   `json:"user_name"`
}

func (r ReqUserInfo) GetUserId() int64 {
	return r.UserId
}

func (r ReqUserInfo) GetUserName() string {
	return r.UserName
}

*/

// GetUserIdFromCtx returns the user ID if it can be obtained from ctx using the UserInfoCtxKey struct if it has a GetUserId() method. Otherwise it returns 0.
func GetUserIdFromCtx(ctx context.Context) int64 {

	m, ok := ctx.Value(UserInfoCtxKey).(interface{ GetUserId() int64 })
	if !ok {
		return 0
	}
	return m.GetUserId()
}

// GetUserNameFromCtx returns the username if it can be obtained from ctx using the UserInfoCtxKey struct if it has a GetUserName() method. Otherwise it returns "Unknown".
func GetUserNameFromCtx(ctx context.Context) string {

	m, ok := ctx.Value(UserInfoCtxKey).(interface{ GetUserName() string })
	if !ok {
		return "Unknown"
	}
	return m.GetUserName()
}
