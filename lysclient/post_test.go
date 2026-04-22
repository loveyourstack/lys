package lysclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type postIn struct {
	Name string `json:"name"`
}

func TestPostToValue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":55,"err_description":""}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		val, err := PostToValue[postIn, int](context.Background(), *srv.Client(), http.MethodPost, srv.URL+"/items", postIn{Name: "a"})
		require.NoError(t, err)
		assert.EqualValues(t, 55, val)
	})

	t.Run("invalid-method", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"succeeded","data":1,"err_description":""}`))
		})

		srv := httptest.NewServer(h)
		defer srv.Close()

		_, err := PostToValue[postIn, int](context.Background(), *srv.Client(), http.MethodGet, srv.URL+"/items", postIn{Name: "a"})
		require.Error(t, err)
		assert.Equal(t, "invalid method: GET", err.Error())
	})
}

func TestPostToValueTester(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":"done","err_description":""}`))
	})

	val, err := PostToValueTester[postIn, string](context.Background(), h, http.MethodPatch, "/items/1", postIn{Name: "z"})
	require.NoError(t, err)
	assert.Equal(t, "done", val)
}

func TestPostArrayToValueTester(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":2,"err_description":""}`))
	})

	val, err := PostArrayToValueTester[postIn, int](context.Background(), h, http.MethodPost, "/import", []postIn{{Name: "x"}, {Name: "y"}})
	require.NoError(t, err)
	assert.EqualValues(t, 2, val)
}

func TestMustPostToValue(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"succeeded","data":99,"err_description":""}`))
	})

	val := MustPostToValue[postIn, int](context.Background(), t, h, http.MethodPut, "/items/1", postIn{Name: "updated"})
	assert.EqualValues(t, 99, val)
}
