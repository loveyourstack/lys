package lys

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
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

func getIdFromReq[idT lyspg.PrimaryKeyType](r *http.Request) (id idT, err error) {

	var zero idT

	// get the id param
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		return zero, ErrIdMissing
	}

	// parse the id param into an idT
	id, err = parseIdByType[idT](idStr)
	if err != nil {
		return zero, ErrIdParseError
	}

	return id, nil
}

// parseIdByType parses a string id (from a request) into an idT.
func parseIdByType[idT lyspg.PrimaryKeyType](idStr string) (idT, error) {

	var zero idT

	switch any(zero).(type) {

	case int64:
		v, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return zero, err
		}
		return any(v).(idT), nil

	case int:
		v, err := strconv.ParseInt(idStr, 10, strconv.IntSize)
		if err != nil {
			return zero, err
		}
		return any(int(v)).(idT), nil

	case string:
		return any(idStr).(idT), nil

	case uuid.UUID:
		v, err := uuid.Parse(idStr)
		if err != nil {
			return zero, err
		}
		return any(v).(idT), nil

	default:
		return zero, fmt.Errorf("unsupported id type %T", zero)
	}
}
