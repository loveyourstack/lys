package lysclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArray(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":[1,2,3],"err_description":""}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		vals, err := GetArray[int](*srv.Client(), srv.URL+"/items")
		require.NoError(t, err)
		assert.EqualValues(t, []int{1, 2, 3}, vals)
	})

	t.Run("status-code-error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"failed","data":[],"err_description":"bad request"}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		_, err := GetArray[int](*srv.Client(), srv.URL+"/items")
		require.Error(t, err)
		assert.Equal(t, fmt.Sprintf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, http.StatusBadRequest, srv.URL+"/items"), err.Error())
	})
}

func TestGetArrayTester(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":["a","b"],"err_description":""}`))
	})

	vals, err := GetArrayTester[string](context.Background(), h, "/items")
	require.NoError(t, err)
	assert.EqualValues(t, []string{"a", "b"}, vals)
}

func TestGetItemResp(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":[{"id":1,"name":"n"}],"metadata":{"count":1,"total_count":1,"total_count_is_estimated":false},"err_description":""}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		resp, err := GetItemResp(*srv.Client(), srv.URL+"/items")
		require.NoError(t, err)
		require.NotNil(t, resp.GetMetadata)
		assert.EqualValues(t, 1, resp.GetMetadata.Count)
		assert.EqualValues(t, 1, len(resp.Data))
	})

	t.Run("tester", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":[{"id":2}],"err_description":""}`))
		})

		resp, err := GetItemRespTester(context.Background(), h, "/items")
		require.NoError(t, err)
		assert.EqualValues(t, 1, len(resp.Data))
	})
}

func TestMustGetHelpers(t *testing.T) {
	t.Run("MustGetArray", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":[10,20],"err_description":""}`))
		})

		vals := MustGetArray[int](context.Background(), t, h, "/arr")
		assert.EqualValues(t, []int{10, 20}, vals)
	})

	t.Run("MustGetItemResp", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":[{"id":3}],"err_description":""}`))
		})

		resp := MustGetItemResp(context.Background(), t, h, "/items")
		assert.EqualValues(t, 1, len(resp.Data))
	})

	t.Run("MustGetValue", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":77,"err_description":""}`))
		})

		val := MustGetValue[int](context.Background(), t, h, "/value")
		assert.EqualValues(t, 77, val)
	})

	t.Run("MustGetFile", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Disposition", "attachment; filename=test.bin")
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("file-content"))
		})

		MustGetFile(context.Background(), t, h, "/file")
	})
}
