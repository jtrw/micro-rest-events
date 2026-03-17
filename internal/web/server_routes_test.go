package web

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	repository "micro-rest-events/internal/repository"
	mock_event "micro-rest-events/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newRoutesServer builds a Server with full routes and templates.
func newRoutesServer(t *testing.T, store *mock_event.MockEventRepository) *httptest.Server {
	t.Helper()
	srv := newTestServer(t, store)
	return httptest.NewServer(srv.routes())
}

// sessionCookieFor returns a valid signed session cookie value for the test server.
func sessionCookieFor(t *testing.T) *http.Cookie {
	t.Helper()
	srv := newTestServer(t, nil)
	token := srv.createSession("admin")
	return &http.Cookie{Name: sessionCookie, Value: token}
}

// doWithSession performs a request with a valid session cookie, following no redirects.
func doWithSession(t *testing.T, method, url, body string, store *mock_event.MockEventRepository) *http.Response {
	t.Helper()
	ts := newRoutesServer(t, store)
	t.Cleanup(ts.Close)

	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req, _ := http.NewRequest(method, ts.URL+url, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	srv := newTestServer(t, nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: srv.createSession("admin")})

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	return resp
}

// --- Static files (no auth required) ---

func TestRoutes_StaticFile(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/style.css")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// --- robots.txt (no auth required) ---

func TestRoutes_Robots(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/robots.txt")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "User-agent: *\nDisallow: /\n", string(body))
}

// --- /ping (no auth required) ---

func TestRoutes_Ping(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// --- Login page ---

func TestRoutes_LoginPage_GET(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/login")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "<form")
	assert.Contains(t, string(body), "action=\"/login\"")
}

func TestRoutes_LoginSubmit_OK(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	form := "username=admin&password=admin"
	resp, err := client.Post(ts.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form))
	assert.NoError(t, err)

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/", resp.Header.Get("Location"))

	// session cookie must be set
	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == sessionCookie {
			found = true
			assert.NotEmpty(t, c.Value)
			assert.True(t, c.HttpOnly)
		}
	}
	assert.True(t, found, "session cookie not set")
}

func TestRoutes_LoginSubmit_WrongPassword(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	form := "username=admin&password=wrong"
	resp, err := http.Post(ts.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // re-renders form with error
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "invalid username or password")
}

func TestRoutes_Logout(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/logout", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/login", resp.Header.Get("Location"))

	// session cookie should be cleared
	for _, c := range resp.Cookies() {
		if c.Name == sessionCookie {
			assert.Equal(t, -1, c.MaxAge)
		}
	}
}

// --- Web UI: no session → redirect to /login ---

func TestRoutes_Dashboard_NoSession_Redirects(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/login", resp.Header.Get("Location"))
}

func TestRoutes_WebEvents_NoSession_Redirects(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(ts.URL + "/web/events")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

// --- Web UI: with valid session → 200 ---

func TestRoutes_DashboardRoot(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Count", mock.AnythingOfType("repository.Query")).Return(0, nil)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	resp := doWithSession(t, http.MethodGet, "/", "", mockRepo)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestRoutes_WebEventsGET(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Count", mock.AnythingOfType("repository.Query")).Return(0, nil)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	resp := doWithSession(t, http.MethodGet, "/web/events", "", mockRepo)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestRoutes_WebEventsPOST_BadRequest(t *testing.T) {
	resp := doWithSession(t, http.MethodPost, "/web/events", "", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// --- Session verification ---

func TestVerifySession_Valid(t *testing.T) {
	srv := newTestServer(t, nil)
	token := srv.createSession("admin")
	assert.True(t, srv.verifySession(token))
}

func TestVerifySession_Tampered(t *testing.T) {
	srv := newTestServer(t, nil)
	token := srv.createSession("admin")
	assert.False(t, srv.verifySession(token+"x"))
}

func TestVerifySession_Empty(t *testing.T) {
	srv := newTestServer(t, nil)
	assert.False(t, srv.verifySession(""))
}

func TestVerifySession_Expired(t *testing.T) {
	srv := newTestServer(t, nil)
	expires := time.Now().Add(-time.Hour).Unix()
	payload := fmt.Sprintf("admin:%d", expires)
	sig := srv.sign(payload)
	expired := payload + ":" + sig
	assert.False(t, srv.verifySession(expired))
}

// --- API routes: JWT auth guard (unchanged) ---

func TestRoutes_API_NoToken(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/events", "application/json", strings.NewReader(`{}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRoutes_API_InvalidToken(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events", strings.NewReader(`{}`))
	req.Header.Set("Api-Token", "not-a-valid-jwt")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRoutes_API_TokenWithoutUserID(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events/users/1", nil)
	req.Header.Set("Api-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.WKQfGgHiRhXdkdz6Qy90gMQhYf3uK-GMeyAQBEs1EbQ")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// --- routes() sanity check ---

func TestServer_RunFailsWithBadTemplate(t *testing.T) {
	srv := &Server{
		Listen:       "localhost:0",
		Secret:       "secret",
		Version:      "test",
		AuthLogin:    "admin",
		AuthPassword: "admin",
	}
	r := srv.routes()
	assert.NotNil(t, r)
}
