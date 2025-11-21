package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareCors(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("blabla blabla"))
		require.NoError(t, err)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()
	{
		req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()

		headers := resp.Header.Get("Access-Control-Allow-Headers")
		assert.False(t, strings.Contains(headers, "X-Requested-With"))
	}

	ts = httptest.NewServer(Cors(handler))
	defer ts.Close()
	{
		req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "http://example.com")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()

		headers := resp.Header.Get("Access-Control-Allow-Headers")
		assert.True(t, strings.Contains(headers, "X-Requested-With"))
	}

	{
		req, err := http.NewRequest("OPTIONS", ts.URL+"/ping", nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		defer resp.Body.Close()
	}
}

func TestMiddlewareCors_AllowedOrigins(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	})

	// Test with specific allowed origins
	os.Setenv("ALLOWED_ORIGINS", "http://allowed.com,http://also-allowed.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	ts := httptest.NewServer(Cors(handler))
	defer ts.Close()

	// Test allowed origin
	{
		req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "http://allowed.com")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()

		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		assert.Equal(t, "http://allowed.com", allowOrigin)
	}

	// Test disallowed origin
	{
		req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "http://notallowed.com")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()

		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		assert.Empty(t, allowOrigin)
	}
}

func TestMiddlewareCors_WildcardOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	})

	// Clear env to use default wildcard
	os.Unsetenv("ALLOWED_ORIGINS")

	ts := httptest.NewServer(Cors(handler))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://any-origin.com")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()

	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	assert.Equal(t, "*", allowOrigin)
}
