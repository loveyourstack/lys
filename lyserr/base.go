package lyserr

// User is an error that is caused by the user and should be reported back to the him rather than being recorded in a log
type User struct {
	Message string // shown to user to help him identify error
}

func (e User) Error() string {
	return e.Message
}
