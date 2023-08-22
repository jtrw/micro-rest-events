package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRest_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv := Server{Listen: "localhost:54009", Version: "v1"}
	err := srv.Run(ctx)
	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestRest_EventCreate(t *testing.T) {
    srv := Server{Listen: "localhost:54009", Version: "v1", Secret: "12345"}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()
    userId := 333
	st := time.Now()
	resp, err := http.Post(ts.URL + "/api/v1/events", "application/json", strings.NewReader(`{"user_id": `+fmt.Sprint(userId)+`,"type": "test"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.True(t, time.Since(st) <= time.Millisecond*30)
}

func TestRest_RobotsCheck(t *testing.T) {
    srv := Server{Listen: "localhost:54009", Version: "v1", Secret: "12345"}

	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/robots.txt")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
    assert.NoError(t, err)
    assert.Equal(t, "User-agent: *\nDisallow: /api/\n", string(body))
}
