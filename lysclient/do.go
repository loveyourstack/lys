package lysclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// DoToValue sends a request without a body to targetUrl. It expects a T in response
func DoToValue[T any](client http.Client, method string, targetUrl string) (val T, err error) {

	// create req
	req, err := http.NewRequest(method, targetUrl, nil)
	if err != nil {
		return val, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do req
	resp, err := client.Do(req)
	if err != nil {
		return val, fmt.Errorf("client.Do failed: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		return val, fmt.Errorf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, resp.StatusCode, targetUrl)
	}

	// read body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return val, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ValueResp[T]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return val, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return val, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// DoToValueTester sends a request without a body to targetUrl using a test handler. It expects a T in response
func DoToValueTester[T any](h http.Handler, method string, targetUrl string) (val T, err error) {

	// create req
	req, err := http.NewRequest(method, targetUrl, nil)
	if err != nil {
		return val, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do req
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// don't check status code: let code progress so that err_description is returned

	// read body
	respBody, err := io.ReadAll(rr.Body)
	if err != nil {
		return val, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ValueResp[T]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		return val, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return val, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// MustDoToValue sends a request without a body to targetUrl using a test handler. It expects a T in response and will fail on any error
func MustDoToValue[T any](t testing.TB, h http.Handler, method string, targetUrl string) (val T) {

	// create req
	req, err := http.NewRequest(method, targetUrl, nil)
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do req
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// check status code
	if rr.Code != http.StatusOK {
		t.Fatalf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, rr.Code, targetUrl)
	}

	// read body
	respBody, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}

	// unmarshal
	var res ValueResp[T]
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// check status
	if res.Status != successStatus {
		t.Fatalf(res.ErrDescription)
	}

	// success
	return res.Data
}
