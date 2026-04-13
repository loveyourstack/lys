package lys

import (
	"context"

	"github.com/loveyourstack/lys/lyspgdb"
)

// UserInfoCtxKey is the key that should be used when binding a user info struct to a request via context
// if you use this key and add a GetUserName() string method to the struct, the username will be included in error logs when using the error handlers in error_handlers.go
const UserInfoCtxKey lyspgdb.ContextKey = "UserInfoKey"

/*

a sample user info struct and method:

type ReqUserInfo struct {
	Roles    []string `json:"roles"`
	UserId   int64    `json:"user_id"`
	UserName string   `json:"user_name"`
}

func (r ReqUserInfo) GetUserName() string {
	return r.UserName
}

*/

// GetUserNameFromCtx returns the username if it can be obtained from ctx using the UserInfoCtxKey struct if it has a GetUserName() method. Otherwise it returns "Unknown"
func GetUserNameFromCtx(ctx context.Context) string {

	m, ok := ctx.Value(UserInfoCtxKey).(interface{ GetUserName() string })
	if !ok {
		return "Unknown"
	}
	return m.GetUserName()
}
