package lys

import (
	"log/slog"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Env (environment) contains objects and options needed by API calls
type Env struct {
	ErrorLog    *slog.Logger
	Validate    *validator.Validate
	GetOptions  GetOptions
	PostOptions PostOptions
}

var parseIdFunc = func(idStr string) (int64, error) {
	return strconv.ParseInt(idStr, 10, 64)
}

var parseUuidFunc = func(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}

// RouteAdderFunc is a function returning a subrouter
type RouteAdderFunc func(r *mux.Router) *mux.Router

// SubRoute contains a Url path and the function returning the subrouter to process that path
type SubRoute struct {
	Url        string
	RouteAdder RouteAdderFunc
}
