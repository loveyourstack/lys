package lys

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// response constants
const (
	// status
	ReqSucceeded string = "succeeded"
	ReqFailed    string = "failed"

	// data
	DataDeleted     string = "deleted"
	DataSoftDeleted string = "soft-deleted"
	DataRestored    string = "restored"
	DataUpdated     string = "updated"
)

// StdResponse is the return type of all API routes
type StdResponse struct {
	Status          string `json:"status"`
	Data            any    `json:"data,omitempty"`
	ErrType         string `json:"err_type,omitempty"`
	ErrDescription  string `json:"err_description,omitempty"`
	ExternalMessage string `json:"external_message,omitempty"` // user-readable messages passed on from 3rd party API calls
}

// RespHeader contains the data in a HTTP reponse header
type RespHeader struct {
	Key   string
	Value string
}

// JsonResponse writes the supplied params to w
func JsonResponse(resp StdResponse, httpStatus int, headers []RespHeader, w http.ResponseWriter) {

	json, err := json.Marshal(resp)
	if err != nil {
		// should never happen
		fmt.Printf("JsonResponse: json.Marshal failed: %s", err.Error())
		return
	}

	// mandatory header
	w.Header().Set("Content-Type", "application/json")

	// add further headers, if any
	for _, v := range headers {
		w.Header().Set(v.Key, v.Value)
	}

	w.WriteHeader(httpStatus)
	w.Write(json)
}
