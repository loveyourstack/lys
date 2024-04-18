package lysclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// GetArray GETs the target Url. It expects an array of T in response
func GetArray[T any](client http.Client, targetUrl string) (arr []T, err error) {

	resp, err := client.Get(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("client.Get failed: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, resp.StatusCode, targetUrl)
	}

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ArrayResp[T]
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return nil, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// GetArrayTester GETs the target Url using a test handler. It expects an array of T in response
func GetArrayTester[T any](h http.Handler, targetUrl string) (arr []T, err error) {

	// create req
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do req
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// don't check status code: let code progress so that err_description is returned

	// read body
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ArrayResp[T]
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return nil, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// GetItems GETs the target Url. It expects an array of items in response
func GetItems(client http.Client, targetUrl string) (items []map[string]any, err error) {

	resp, err := client.Get(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("client.Get failed: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, resp.StatusCode, targetUrl)
	}

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ItemAResp
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return nil, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// GetItemsTester GETs the target Url using a test handler. It expects an array of items in response
func GetItemsTester(h http.Handler, targetUrl string) (items []map[string]any, err error) {

	// create req
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// do req
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// don't check status code: let code progress so that err_description is returned

	// read body
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// unmarshal
	var res ItemAResp
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// check status
	if res.Status != successStatus {
		return nil, fmt.Errorf(res.ErrDescription)
	}

	// success
	return res.Data, nil
}

// MustGetArray GETs the target Url using a test handler. It expects an array of T in response and will fail on any error
func MustGetArray[T any](t testing.TB, h http.Handler, targetUrl string) (arr []T) {

	// create req
	req, err := http.NewRequest("GET", targetUrl, nil)
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
	var res ArrayResp[T]
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

// MustGetItems GETs the target Url using a test handler. It expects an array of items in response and will fail on any error
func MustGetItems(t testing.TB, h http.Handler, targetUrl string) (items []map[string]any) {

	// create req
	req, err := http.NewRequest("GET", targetUrl, nil)
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
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}

	// unmarshal
	var res ItemAResp
	err = json.Unmarshal(body, &res)
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
