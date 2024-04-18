package lys

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// Env (environment) contains objects and options needed by API calls
type Env struct {
	ErrorLog    *slog.Logger
	Validate    *validator.Validate
	GetOptions  GetOptions
	PostOptions PostOptions
}

// RouteAdderFunc is a function returning a subrouter
type RouteAdderFunc func(r *mux.Router) *mux.Router

// SubRoute contains a Url path and the function returning the subrouter to process that path
type SubRoute struct {
	Url        string
	RouteAdder RouteAdderFunc
}
