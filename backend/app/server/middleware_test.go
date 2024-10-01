package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
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
