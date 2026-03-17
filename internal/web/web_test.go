package web

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	repository "micro-rest-events/internal/repository"
	mock_event "micro-rest-events/internal/repository/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newTestServer builds a Server with parsed templates and a mock store.
func newTestServer(t *testing.T, store repository.StoreProviderInterface) *Server {
	t.Helper()
	tmpl, err := template.New("").ParseFS(embedFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	return &Server{
		Listen:        "localhost:0",
		Secret:        "secret",
		Version:       "test",
		StoreProvider: store,
		tmpl:          tmpl,
	}
}

// --- dashboard ---

func TestDashboard_OK(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{{Uuid: "a1b2c3d4-uuid", Status: "new"}}, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.dashboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "<!DOCTYPE html>")
	mockRepo.AssertExpectations(t)
}

func TestDashboard_FilterByUserID(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAllByUserId", "user-1", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/?user_id=user-1", nil)
	rr := httptest.NewRecorder()
	srv.dashboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestDashboard_StoreError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, fmt.Errorf("db error"))

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	srv.dashboard(rr, req)

	// error from store is swallowed in loadDashboardData — dashboard still renders
	assert.Equal(t, http.StatusOK, rr.Code)
}

// --- eventsTable ---

func TestEventsTable_OK(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/web/events", nil)
	rr := httptest.NewRecorder()
	srv.eventsTable(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestEventsTable_WithEvents(t *testing.T) {
	events := []repository.Event{
		{Uuid: "aaaaaaaa-uuid-1", UserId: "u1", Type: "alert", Status: "new", Caption: "cap"},
		{Uuid: "bbbbbbbb-uuid-2", UserId: "u1", Type: "info", Status: "done", IsSeen: true},
	}
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).Return(events, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/web/events", nil)
	rr := httptest.NewRecorder()
	srv.eventsTable(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "aaaaaaaa")
	assert.Contains(t, body, "bbbbbbbb")
	mockRepo.AssertExpectations(t)
}

// --- createEvent ---

func TestCreateEvent_OK(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)
	// after Create, loadDashboardData sees user_id=u1 in the parsed form → GetAllByUserId
	mockRepo.On("GetAllByUserId", "u1", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	srv := newTestServer(t, mockRepo)

	form := url.Values{"user_id": {"u1"}, "type": {"alert"}, "caption": {"cap"}, "body": {"msg"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.createEvent(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestCreateEvent_MissingUserID(t *testing.T) {
	srv := newTestServer(t, nil)

	form := url.Values{"type": {"alert"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.createEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateEvent_MissingType(t *testing.T) {
	srv := newTestServer(t, nil)

	form := url.Values{"user_id": {"u1"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.createEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateEvent_StoreError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).
		Return(fmt.Errorf("db error"))

	srv := newTestServer(t, mockRepo)

	form := url.Values{"user_id": {"u1"}, "type": {"alert"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	srv.createEvent(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockRepo.AssertExpectations(t)
}

// --- changeStatus ---

func TestChangeStatus_OK(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "test-uuid", mock.AnythingOfType("repository.Event")).
		Return(int64(1), nil)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	srv := newTestServer(t, mockRepo)
	r := chi.NewRouter()
	r.Post("/web/events/{uuid}/status", srv.changeStatus)

	form := url.Values{"status": {"done"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events/test-uuid/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestChangeStatus_MissingStatus(t *testing.T) {
	srv := newTestServer(t, nil)
	r := chi.NewRouter()
	r.Post("/web/events/{uuid}/status", srv.changeStatus)

	req := httptest.NewRequest(http.MethodPost, "/web/events/test-uuid/status", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestChangeStatus_StoreError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "test-uuid", mock.AnythingOfType("repository.Event")).
		Return(int64(0), fmt.Errorf("db error"))

	srv := newTestServer(t, mockRepo)
	r := chi.NewRouter()
	r.Post("/web/events/{uuid}/status", srv.changeStatus)

	form := url.Values{"status": {"done"}}
	req := httptest.NewRequest(http.MethodPost, "/web/events/test-uuid/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockRepo.AssertExpectations(t)
}

// --- markSeen ---

func TestMarkSeen_OK(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "test-uuid").Return(int64(1), nil)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, nil)

	srv := newTestServer(t, mockRepo)
	r := chi.NewRouter()
	r.Post("/web/events/{uuid}/seen", srv.markSeen)

	req := httptest.NewRequest(http.MethodPost, "/web/events/test-uuid/seen", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestMarkSeen_StoreError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "test-uuid").Return(int64(0), fmt.Errorf("db error"))

	srv := newTestServer(t, mockRepo)
	r := chi.NewRouter()
	r.Post("/web/events/{uuid}/seen", srv.markSeen)

	req := httptest.NewRequest(http.MethodPost, "/web/events/test-uuid/seen", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockRepo.AssertExpectations(t)
}

// --- loadDashboardData ---

func TestLoadDashboardData_Counts(t *testing.T) {
	events := []repository.Event{
		{Status: "new", IsSeen: false},
		{Status: "done", IsSeen: true},
		{Status: "new", IsSeen: true},
	}
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).Return(events, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	data, err := srv.loadDashboardData(req)

	assert.NoError(t, err)
	assert.Equal(t, 3, data.Total)
	assert.Equal(t, 2, data.New)
	assert.Equal(t, 2, data.Seen)
	mockRepo.AssertExpectations(t)
}

func TestLoadDashboardData_WithUserID(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAllByUserId", "u42", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{{Uuid: "x", Status: "new"}}, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/?user_id=u42", nil)

	data, err := srv.loadDashboardData(req)

	assert.NoError(t, err)
	assert.Equal(t, "u42", data.UserID)
	assert.Equal(t, 1, data.Total)
	mockRepo.AssertExpectations(t)
}

func TestLoadDashboardData_WithStatusFilter(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", repository.Query{Statuses: []string{"done"}, DateFrom: ""}).
		Return([]repository.Event{{Status: "done"}}, nil)

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/?status=done", nil)

	data, err := srv.loadDashboardData(req)

	assert.NoError(t, err)
	assert.Equal(t, "done", data.Status)
	assert.Equal(t, 1, data.Total)
	mockRepo.AssertExpectations(t)
}

func TestLoadDashboardData_StoreErrorReturnsEmptySlice(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("GetAll", mock.AnythingOfType("repository.Query")).
		Return([]repository.Event{}, fmt.Errorf("db error"))

	srv := newTestServer(t, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	data, err := srv.loadDashboardData(req)

	assert.NoError(t, err)
	assert.Empty(t, data.Events)
	assert.Equal(t, 0, data.Total)
}
