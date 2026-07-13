package lys

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
	Status         string       `json:"status"`
	Data           any          `json:"data,omitempty"`
	GetMetadata    *GetMetadata `json:"metadata,omitempty"` // only used for GET many
	ErrDescription string       `json:"err_description,omitempty"`
}

// FileResponse opens the supplied file and streams it to w as a file
func FileResponse(filePath, outputFileName string, remove bool, w http.ResponseWriter) {

	file, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "FileResponse: os.Open failed: %s", err.Error())
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "FileResponse: file.Close failed: %s", err.Error())
		}
	}()

	if remove {
		defer func() {
			if err = os.Remove(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "FileResponse: os.Remove failed: %s", err.Error())
			}
		}()
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", outputFileName))

	if _, err := io.Copy(w, file); err != nil {
		fmt.Fprintf(os.Stderr, "FileResponse: io.Copy failed: %s", err.Error())
	}
}

// JsonResponse marshals the supplied StdResponse to json and writes it to w
func JsonResponse(resp StdResponse, httpStatus int, w http.ResponseWriter) {

	b, err := json.Marshal(resp)
	if err != nil {
		// should never happen
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "JsonResponse: json.Marshal failed: %s", err.Error())
		return
	}

	// mandatory header
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(httpStatus)
	if _, err := w.Write(b); err != nil {
		fmt.Fprintf(os.Stderr, "JsonResponse: w.Write failed: %s", err.Error())
	}
}

// StatusWriter is a wrapper around http.ResponseWriter that captures the status code and number of bytes written in the response.
// It implements http.Flusher and http.Hijacker so that it can also write websocket responses.
type StatusWriter struct {
	http.ResponseWriter
	Status int
	Bytes  int
}

func (sw *StatusWriter) Flush() {
	if fl, ok := sw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (sw *StatusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := sw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

func (sw *StatusWriter) Unwrap() http.ResponseWriter {
	return sw.ResponseWriter
}

func (sw *StatusWriter) WriteHeader(code int) {
	sw.Status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *StatusWriter) Write(b []byte) (int, error) {
	if sw.Status == 0 {
		sw.Status = http.StatusOK
	}
	n, err := sw.ResponseWriter.Write(b)
	sw.Bytes += n
	return n, err
}
