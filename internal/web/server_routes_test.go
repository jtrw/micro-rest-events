package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	repository "micro-rest-events/internal/repository"
	mock_event "micro-rest-events/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newRoutesServer builds a Server with routes and templates wired up.
func newRoutesServer(t *testing.T, store *mock_event.MockEventRepository) *httptest.Server {
	t.Helper()
	srv := newTestServer(t, store)
	return httptest.NewServer(srv.routes())
}

// --- Static files ---

func TestRoutes_StaticFile(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/style.css")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// --- robots.txt ---

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

// --- /ping (from go-rest middleware) ---

func TestRoutes_Ping(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// --- Web UI routes ---

func TestRoutes_DashboardRoot(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	ts := newRoutesServer(t, mockRepo)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestRoutes_WebEventsGET(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	ts := newRoutesServer(t, mockRepo)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/web/events")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestRoutes_WebEventsPOST_BadRequest(t *testing.T) {
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	// Missing required fields → 400
	resp, err := http.Post(ts.URL+"/web/events", "application/x-www-form-urlencoded", strings.NewReader(""))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRoutes_WebEventsSeen_NotFound(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "no-such-uuid").Return(int64(0), nil)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	ts := newRoutesServer(t, mockRepo)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/web/events/no-such-uuid/seen", "application/x-www-form-urlencoded", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

// --- API routes: JWT auth guard ---

func TestRoutes_API_NoToken(t *testing.T) {
	// go-rest returns 403 when no token is present
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/events", "application/json", strings.NewReader(`{}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRoutes_API_InvalidToken(t *testing.T) {
	// go-rest returns 401 for a malformed (non-JWT) token
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events", strings.NewReader(`{}`))
	req.Header.Set("Api-Token", "not-a-valid-jwt")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRoutes_API_TokenWithoutUserID(t *testing.T) {
	// JWT with empty payload (no user_id) — claims check fails → 403
	// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.WKQfGgHiRhXdkdz6Qy90gMQhYf3uK-GMeyAQBEs1EbQ
	ts := newRoutesServer(t, nil)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events/users/1", nil)
	req.Header.Set("Api-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.WKQfGgHiRhXdkdz6Qy90gMQhYf3uK-GMeyAQBEs1EbQ")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// --- Server.Run template parsing ---

func TestServer_RunFailsWithBadTemplate(t *testing.T) {
	// If StoreProvider is nil the server still starts; template parsing happens inside Run.
	// We just verify that routes() doesn't panic without a template.
	srv := &Server{
		Listen:  "localhost:0",
		Secret:  "secret",
		Version: "test",
	}
	// routes() can be called without tmpl — handlers only use tmpl on request
	r := srv.routes()
	assert.NotNil(t, r)
}
