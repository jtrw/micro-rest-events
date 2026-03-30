package web

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
)

// --- Run ---

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv := Server{Listen: "localhost:54009", Version: "v1"}
	err := srv.Run(ctx)
	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestRun_BadAddress(t *testing.T) {
	srv := Server{Listen: "localhost:99999", Version: "v1"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := srv.Run(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server failed")
}

// --- routes ---

func TestRoutes_NotNil(t *testing.T) {
	srv := &Server{Listen: "localhost:0", Secret: "s", Version: "v", AuthLogin: "admin", AuthPassword: "admin"}
	r := srv.routes()
	assert.NotNil(t, r)
}

// --- AuthMiddleware ---

func TestAuthMiddleware_Disabled_AlwaysPasses(t *testing.T) {
	srv := newTestServer(t, nil)
	srv.AuthPassword = "" // auth disabled

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.AuthMiddleware(next).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_PublicPaths_Passthrough(t *testing.T) {
	srv := newTestServer(t, nil)

	publicPaths := []string{
		"/login",
		"/logout",
		"/static/style.css",
		"/api/v1/events",
		"/ping",
		"/robots.txt",
	}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			srv.AuthMiddleware(next).ServeHTTP(rr, req)
			assert.True(t, called, "expected passthrough for %s", path)
		})
	}
}

func TestAuthMiddleware_NoCookie_Redirects(t *testing.T) {
	srv := newTestServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.AuthMiddleware(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/login", rr.Header().Get("Location"))
}

func TestAuthMiddleware_InvalidCookie_Redirects(t *testing.T) {
	srv := newTestServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "tampered:value:badsig"})
	rr := httptest.NewRecorder()
	srv.AuthMiddleware(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/login", rr.Header().Get("Location"))
}

func TestAuthMiddleware_ValidCookie_Passthrough(t *testing.T) {
	srv := newTestServer(t, nil)
	token := srv.createSession("admin")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	rr := httptest.NewRecorder()
	srv.AuthMiddleware(next).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_ExpiredCookie_Redirects(t *testing.T) {
	srv := newTestServer(t, nil)

	expiredPayload := fmt.Sprintf("admin:%d", time.Now().Add(-time.Hour).Unix())
	token := expiredPayload + ":" + srv.sign(expiredPayload)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	rr := httptest.NewRecorder()
	srv.AuthMiddleware(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/login", rr.Header().Get("Location"))
}

// --- handleLoginPage ---

func TestHandleLoginPage_AuthDisabled_Redirects(t *testing.T) {
	srv := newTestServer(t, nil)
	srv.AuthPassword = ""

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()
	srv.handleLoginPage(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/", rr.Header().Get("Location"))
}

func TestHandleLoginPage_AuthEnabled_RendersForm(t *testing.T) {
	srv := newTestServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()
	srv.handleLoginPage(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "<form")
}

// --- handleLogout ---

func TestHandleLogout_ClearsCookieAndRedirects(t *testing.T) {
	srv := newTestServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rr := httptest.NewRecorder()
	srv.handleLogout(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/login", rr.Header().Get("Location"))

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == sessionCookie {
			found = true
			assert.Equal(t, -1, c.MaxAge)
		}
	}
	assert.True(t, found, "session cookie not cleared")
}

// --- handleLoginSubmit ---

func TestHandleLoginSubmit_ValidCredentials(t *testing.T) {
	srv := newTestServer(t, nil)

	form := "username=admin&password=admin"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.handleLoginSubmit(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "/", rr.Header().Get("Location"))

	var found bool
	for _, c := range rr.Result().Cookies() {
		if c.Name == sessionCookie {
			found = true
			assert.NotEmpty(t, c.Value)
			assert.True(t, c.HttpOnly)
		}
	}
	assert.True(t, found, "session cookie not set")
}

func TestHandleLoginSubmit_WrongPassword(t *testing.T) {
	srv := newTestServer(t, nil)

	form := "username=admin&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.handleLoginSubmit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid username or password")
}

func TestHandleLoginSubmit_WrongUsername(t *testing.T) {
	srv := newTestServer(t, nil)

	form := "username=hacker&password=admin"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.handleLoginSubmit(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid username or password")
}

// --- sign / createSession / verifySession ---

func TestSign_Deterministic(t *testing.T) {
	srv := newTestServer(t, nil)
	assert.Equal(t, srv.sign("payload"), srv.sign("payload"))
}

func TestSign_DifferentSecrets(t *testing.T) {
	srv1 := newTestServer(t, nil)
	srv2 := &Server{Secret: "other-secret"}
	assert.NotEqual(t, srv1.sign("payload"), srv2.sign("payload"))
}

func TestCreateSession_VerifyRoundtrip(t *testing.T) {
	srv := newTestServer(t, nil)
	token := srv.createSession("admin")
	assert.True(t, srv.verifySession(token))
}

func TestVerifySession_WrongSecret(t *testing.T) {
	srv1 := newTestServer(t, nil)
	srv2 := &Server{Secret: "other"}
	token := srv1.createSession("admin")
	assert.False(t, srv2.verifySession(token))
}

// --- robots.txt ---

func TestRobotsCheck(t *testing.T) {
	srv := &Server{Listen: "localhost:54009", Version: "v1", Secret: "12345"}
	ts := httptest.NewServer(srv.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/robots.txt")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "User-agent: *\nDisallow: /\n", string(body))
}

// --- API with real DB ---

const validJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozMzN9.DblZ_sMWqugXaEO4v1q1rrprOdH1YANKI2Q1NWeE9mg"
const apiSecret = "1234567890"

func newAPIServer(t *testing.T) (*httptest.Server, repository.StoreProviderInterface) {
	t.Helper()
	store, err := repository.NewStoreProvider("file:///tmp/test_server_api.db")
	require.NoError(t, err)
	srv := &Server{
		Listen:        "127.0.0.1:0",
		Secret:        apiSecret,
		Version:       "test",
		StoreProvider: store,
	}
	ts := httptest.NewServer(srv.routes())
	t.Cleanup(ts.Close)
	return ts, store
}

func TestAPI_CreateEvent(t *testing.T) {
	ts, _ := newAPIServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
		strings.NewReader(`{"user_id":"333","type":"test"}`))
	req.Header.Set("Api-Token", validJWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
	assert.NotEmpty(t, body["uuid"])
}

func TestAPI_CreateBatchEvents(t *testing.T) {
	ts, _ := newAPIServer(t)

	payload := `{"type":"alert","users":["user1","user2","user3"]}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events/batch",
		strings.NewReader(payload))
	req.Header.Set("Api-Token", validJWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
}

func TestAPI_CreateEvent_MissingType(t *testing.T) {
	ts, _ := newAPIServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
		strings.NewReader(`{"user_id":"333"}`))
	req.Header.Set("Api-Token", validJWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAPI_GetEventsByUserId(t *testing.T) {
	ts, _ := newAPIServer(t)

	// create an event first
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
		strings.NewReader(`{"user_id":"u-get-test","type":"info"}`))
	req.Header.Set("Api-Token", validJWT)
	http.DefaultClient.Do(req)

	// get events
	req2, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events/users/u-get-test", nil)
	req2.Header.Set("Api-Token", validJWT)

	resp, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
}

func TestAPI_ChangeEvent(t *testing.T) {
	ts, _ := newAPIServer(t)

	// create
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
		strings.NewReader(`{"user_id":"u-change","type":"info"}`))
	req.Header.Set("Api-Token", validJWT)
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	uuid := created["uuid"].(string)

	// change status
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events/"+uuid,
		strings.NewReader(`{"status":"done"}`))
	req2.Header.Set("Api-Token", validJWT)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestAPI_SetSeen(t *testing.T) {
	ts, _ := newAPIServer(t)

	// create
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
		strings.NewReader(`{"user_id":"u-seen","type":"info"}`))
	req.Header.Set("Api-Token", validJWT)
	resp, _ := http.DefaultClient.Do(req)
	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	uuid := created["uuid"].(string)

	// mark seen
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events/"+uuid+"/seen", nil)
	req2.Header.Set("Api-Token", validJWT)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestAPI_ChangeBatchEvents(t *testing.T) {
	ts, _ := newAPIServer(t)

	// create two events
	uuids := make([]string, 2)
	for i, uid := range []string{"u-batch-1", "u-batch-2"} {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events",
			strings.NewReader(`{"user_id":"`+uid+`","type":"info"}`))
		req.Header.Set("Api-Token", validJWT)
		resp, _ := http.DefaultClient.Do(req)
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		uuids[i] = body["uuid"].(string)
	}

	payload := `{"uuids":["` + uuids[0] + `","` + uuids[1] + `"],"status":"done"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events/change/batch",
		strings.NewReader(payload))
	req.Header.Set("Api-Token", validJWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, float64(2), body["count"])
}

func TestAPI_JwtNoToken(t *testing.T) {
	ts, _ := newAPIServer(t)

	resp, err := http.Post(ts.URL+"/api/v1/events", "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAPI_JwtInvalidToken(t *testing.T) {
	ts, _ := newAPIServer(t)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/events", strings.NewReader(`{}`))
	req.Header.Set("Api-Token", "not-a-jwt")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// --- mock-based API tests ---

func TestAPI_GetEventsByUserId_Mock(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAllByUserId", "u99", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{{Uuid: "aaaaaaaa-1111", UserId: "u99", Status: "new"}}, nil)

	srv := &Server{
		Listen:        "localhost:0",
		Secret:        apiSecret,
		Version:       "test",
		AuthLogin:     "admin",
		AuthPassword:  "admin",
		StoreProvider: mockRepo,
	}
	ts := httptest.NewServer(srv.routes())
	t.Cleanup(ts.Close)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/events/users/u99", nil)
	req.Header.Set("Api-Token", validJWT)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}
