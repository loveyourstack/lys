package lys

// validation user errors
const (
	ErrDescBodyMissing        string = "request body missing"
	ErrDescIdMissing          string = "id missing"
	ErrDescIdNotAUuid         string = "id not a uuid"
	ErrDescIdNotAnInteger     string = "id not an integer"
	ErrDescIdNotUnique        string = "id not unique" // the handling func was expecting id to be unique, but it is not
	ErrDescInvalidContentType string = "content type must be application/json"
	ErrDescInvalidId          string = "invalid id" // the Id sent is not present in the relevant table
	ErrDescRouteNotFound      string = "route not found"
	ErrDescUserInfoMissing    string = "userInfo missing"  // failed to get ReqUserInfo from context
	ErrDescPermissionDenied   string = "permission denied" // authorization failed
)
