package lysclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// MustPostToValue sends a POST/PUT/PATCH request to targetUrl using a test handler with an inT as the body. It expects an outT in response and will fail on any error
func MustPostToValue[inT, outT any](t testing.TB, h http.Handler, method string, targetUrl string, item inT) (val outT) {

	// check method
	if !slices.Contains([]string{"POST", "PUT", "PATCH"}, method) {
		t.Fatalf("invalid method: %s", method)
	}

	// marshal
	reqBody, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// create req
	req, err := http.NewRequest(method, targetUrl, bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do request
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// read body
	respBody, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}

	// unmarshal
	var res ValueResp[outT]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// check status
	if res.Status != successStatus {
		t.Fatalf(res.ErrDescription)
	}

	return res.Data
}

// PostToValue sends a POST/PUT/PATCH request to targetUrl with an inT as the body. It expects an outT in response
func PostToValue[inT, outT any](client http.Client, method string, targetUrl string, item inT) (val outT, err error) {

	// check method
	if !slices.Contains([]string{"POST", "PUT", "PATCH"}, method) {
		return val, fmt.Errorf("invalid method: %s", method)
	}

	// marshal
	reqBody, err := json.Marshal(item)
	if err != nil {
		return val, fmt.Errorf("json.Marshal failed: %w", err)
	}

	// create req
	req, err := http.NewRequest(method, targetUrl, bytes.NewReader(reqBody))
	if err != nil {
		return val, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do request
	resp, err := client.Do(req)
	if err != nil {
		return val, fmt.Errorf("client.Do failed: %w", err)
	}
	defer resp.Body.Close()

	// read body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return val, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ValueResp[outT]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return val, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return val, fmt.Errorf(res.ErrDescription)
	}

	return res.Data, nil
}

// PostToValueTester sends a POST/PUT/PATCH request to targetUrl using a test handler with an inT as the body. It expects an outT in response
func PostToValueTester[inT, outT any](h http.Handler, method string, targetUrl string, item inT) (val outT, err error) {

	// check method
	if !slices.Contains([]string{"POST", "PUT", "PATCH"}, method) {
		return val, fmt.Errorf("invalid method: %s", method)
	}

	// marshal
	reqBody, err := json.Marshal(item)
	if err != nil {
		return val, fmt.Errorf("json.Marshal failed: %w", err)
	}

	// create req
	req, err := http.NewRequest(method, targetUrl, bytes.NewReader(reqBody))
	if err != nil {
		return val, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do request
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// read body
	respBody, err := io.ReadAll(rr.Body)
	if err != nil {
		return val, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ValueResp[outT]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return val, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return val, fmt.Errorf(res.ErrDescription)
	}

	return res.Data, nil
}
