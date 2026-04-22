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

func TestDoToValue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":123,"err_description":""}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		val, err := DoToValue[int](context.Background(), *srv.Client(), http.MethodGet, srv.URL+"/value")
		require.NoError(t, err)
		assert.EqualValues(t, 123, val)
	})

	t.Run("status-code-error", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"failed","data":0,"err_description":"boom"}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		_, err := DoToValue[int](context.Background(), *srv.Client(), http.MethodGet, srv.URL+"/value")
		require.Error(t, err)
		assert.Equal(t, fmt.Sprintf("expected statusCode: %d, got: %d for Url: %s", http.StatusOK, http.StatusInternalServerError, srv.URL+"/value"), err.Error())
	})
}

func TestDoToValueTester(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":321,"err_description":""}`))
	})

	val, err := DoToValueTester[int](context.Background(), h, http.MethodGet, "/value")
	require.NoError(t, err)
	assert.EqualValues(t, 321, val)
}

func TestMustDoToValue(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":"ok","err_description":""}`))
	})

	val := MustDoToValue[string](context.Background(), t, h, http.MethodGet, "/value")
	assert.Equal(t, "ok", val)
}
