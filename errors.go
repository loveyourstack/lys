package lys

import (
	"net/http"

	"github.com/loveyourstack/lys/lyserr"
)

// validation user errors
var (

	// bad requests (default status)
	ErrBodyMissing        = lyserr.User{Message: "request body missing"}
	ErrIdMissing          = lyserr.User{Message: "id missing"}
	ErrIdNotAUuid         = lyserr.User{Message: "id not a uuid"}
	ErrIdNotAnInteger     = lyserr.User{Message: "id not an integer"}
	ErrIdNotUnique        = lyserr.User{Message: "id not unique"} // the handling func was expecting id to be unique, but it is not
	ErrInvalidContentType = lyserr.User{Message: "content type must be application/json"}
	ErrInvalidId          = lyserr.User{Message: "invalid id"} // the id sent is not present in the relevant table
	ErrInvalidJson        = lyserr.User{Message: "invalid json"}
	ErrNoAssignments      = lyserr.User{Message: "no assignments found"} // for patch reqs where assignmentMap is expected
	ErrRouteNotFound      = lyserr.User{Message: "route not found"}

	// forbidden
	ErrPermissionDenied = lyserr.User{Message: "permission denied", StatusCode: http.StatusForbidden} // authorization failed
	ErrUserInfoMissing  = lyserr.User{Message: "userInfo missing", StatusCode: http.StatusForbidden}  // failed to get ReqUserInfo from context
)
