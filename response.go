package lys

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/loveyourstack/lys/lystype"
)

// response constants
const (
	// status
	ReqSucceeded string = "succeeded"
	ReqFailed    string = "failed"

	// data
	DataArchived string = "archived"
	DataDeleted  string = "deleted"
	DataRestored string = "restored"
	DataUpdated  string = "updated"
)

type GetMetadata struct {
	Count                 int   `json:"count"`
	TotalCount            int64 `json:"total_count"`
	TotalCountIsEstimated bool  `json:"total_count_is_estimated"`
}

// StdResponse is the return type of all API routes
type StdResponse struct {
	Status          string            `json:"status"`
	Data            any               `json:"data,omitempty"`
	GetMetadata     *GetMetadata      `json:"metadata,omitempty"`     // only used for GET many
	LastSyncAt      *lystype.Datetime `json:"last_sync_at,omitempty"` // if the data was synced from external source: the last sync timestamp
	ErrType         string            `json:"err_type,omitempty"`
	ErrDescription  string            `json:"err_description,omitempty"`
	ExternalMessage string            `json:"external_message,omitempty"` // user-readable messages passed on from 3rd party API calls
}

// FileResponse opens the supplied file and streams it to w as a file
func FileResponse(filePath, outputFileName string, remove bool, w http.ResponseWriter) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("FileResponse: os.Open failed: %s", err.Error())
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			fmt.Printf("FileResponse: file.Close failed: %s", err.Error())
		}
	}()

	if remove {
		defer func() {
			if err = os.Remove(filePath); err != nil {
				fmt.Printf("FileResponse: os.Remove failed: %s", err.Error())
			}
		}()
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", outputFileName))

	io.Copy(w, file)
}

// JsonResponse marshals the supplied StdResponse to json and writes it to w
func JsonResponse(resp StdResponse, httpStatus int, w http.ResponseWriter) {

	json, err := json.Marshal(resp)
	if err != nil {
		// should never happen
		fmt.Printf("JsonResponse: json.Marshal failed: %s", err.Error())
		return
	}

	// mandatory header
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(httpStatus)
	w.Write(json)
}
