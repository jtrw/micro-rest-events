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

func TestAuthenticationJwtMiddleware_StatusUnauthorized(t *testing.T) {
    var jwtToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.WKQfGgHiRhXdkdz6Qy90gMQhYf3uK-GMeyAQBEs1EbQ"
	srv := Server{
		Listen:     "127.0.0.1:8080",
		Secret:     "1234567890",
		Version:    "1.0",
	}

	r := srv.routes()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Створення запиту
	req, err := http.NewRequest("GET", ts.URL+"/api/v1/events/users/1", nil)
	assert.NoError(t, err)

	// Додавання заголовка з JWT токеном (замініть на ваш токен)
	req.Header.Add("Api-Token", jwtToken)

	// Відправка запиту
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) // Перевірка коду статусу
}
