package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	repository "micro-rest-events/internal/repository"
	mock_event "micro-rest-events/internal/repository/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOnCreateEvent(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)

	h := Handler{StoreProvider: mockRepo}

	payload := map[string]interface{}{
		"type":    "test_type",
		"user_id": "123",
		"caption": "test_caption",
		"body":    "test_body",
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	h.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnCreateBatchEvent(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)

	h := Handler{StoreProvider: mockRepo}

	payload := map[string]interface{}{
		"type":    "test_type",
		"users":   []string{"123-1234-11111", "123-1234-22222"},
		"caption": "test_caption",
		"body":    "test_body",
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	h.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnCreateBatchEvent_TypeNotFound(t *testing.T) {
	h := Handler{}

	payload := map[string]interface{}{
		"users":   []string{"123-1234-11111", "123-1234-22222"},
		"caption": "test_caption",
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader(payloadBytes))

	rr := httptest.NewRecorder()
	h.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateBatchEvent_BadData(t *testing.T) {
	h := Handler{}
	req, _ := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader([]byte("bla")))

	rr := httptest.NewRecorder()
	h.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_EmptyPayload(t *testing.T) {
	h := Handler{}

	payload := map[string]interface{}{"type": "", "user_id": nil}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(payloadBytes))

	rr := httptest.NewRecorder()
	h.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_BadJsonRequest(t *testing.T) {
	h := Handler{}
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewReader([]byte("Bad json request")))

	rr := httptest.NewRecorder()
	h.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_RepositoryError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(fmt.Errorf("Unexpected error"))

	h := Handler{StoreProvider: mockRepo}

	payload := map[string]interface{}{
		"type":    "test_type",
		"user_id": "123",
		"caption": "test_caption",
		"body":    "test_body",
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(payloadBytes))

	rr := httptest.NewRecorder()
	h.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnGetEventsByUserId(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockEvent := []repository.Event{{Uuid: "test_uuid", UserId: "123", Type: "test_type", Status: "new"}}
	mockRepo.On("GetAllByUserId", "123", mock.AnythingOfType("repository.Query")).Return(mockEvent, nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Get("/api/v1/events/users/{id}", h.OnGetEventsByUserId)

	req, _ := http.NewRequest("GET", "/api/v1/events/users/123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnGetEventsByUserId_NotFound(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	var mockEvent []repository.Event
	mockRepo.On("GetAllByUserId", "123", mock.AnythingOfType("repository.Query")).Return(mockEvent, fmt.Errorf("Event not found"))

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Get("/api/v1/events/users/{id}", h.OnGetEventsByUserId)

	req, _ := http.NewRequest("GET", "/api/v1/events/users/123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnChangeEvent(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "test_uuid", mock.AnythingOfType("repository.Event")).Return(int64(1), nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(`{"status": "new_status", "message": "new_message"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnChangeBatchEvents(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "123-1234-11111", mock.AnythingOfType("repository.Event")).Return(int64(1), nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

	payload := map[string]interface{}{"status": "delivered", "uuids": []string{"123-1234-11111"}}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/events/change/batch", bytes.NewReader(payloadBytes))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnChangeBatchEvents_NotFoundUuids(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

	payload := map[string]interface{}{"status": "delivered"}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/events/change/batch", bytes.NewReader(payloadBytes))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeBatchEvents_BadJson(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

	req, _ := http.NewRequest("POST", "/api/v1/events/change/batch", strings.NewReader("Bad Json"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeEvent_BadPayload(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader("Bad payload"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeEvent_StatusNotFound(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(`{"message": "new_message"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var requestData JSON
	respBody, err := io.ReadAll(rr.Body)
	err = json.Unmarshal(respBody, &requestData)
	assert.NoError(t, err)
	assert.Equal(t, requestData["message"], "Status is required")
}

func TestOnChangeEvent_NotFound(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "test_uuid", mock.AnythingOfType("repository.Event")).Return(int64(0), nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(`{"status": "new_status"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnChangeEvent_RepositoryReturnError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeStatus", "test_uuid", mock.AnythingOfType("repository.Event")).Return(int64(0), fmt.Errorf("Event not found"))

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(`{"status": "new_status"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnSetSeen(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(1), nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnSetSeenNotFound(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(0), nil)

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnSetSeenError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(0), fmt.Errorf("Some error"))

	h := Handler{StoreProvider: mockRepo}
	r := chi.NewRouter()
	r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

	req, _ := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnChangeBatchEvents_InvalidUuidsFormat(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

	req, _ := http.NewRequest("POST", "/api/v1/events/change/batch", strings.NewReader(`{"status": "delivered", "uuids": "not-an-array"}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseData JSON
	err := json.Unmarshal(rr.Body.Bytes(), &responseData)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid uuids format", responseData["message"])
}

func TestOnChangeBatchEvents_InvalidStatusFormat(t *testing.T) {
	h := Handler{}
	r := chi.NewRouter()
	r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

	req, _ := http.NewRequest("POST", "/api/v1/events/change/batch", strings.NewReader(`{"status": 123, "uuids": ["uuid1", "uuid2"]}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var responseData JSON
	err := json.Unmarshal(rr.Body.Bytes(), &responseData)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid status format", responseData["message"])
}

func TestOnCreateEvent_InvalidUserIdFormat(t *testing.T) {
	h := Handler{}
	req, _ := http.NewRequest("POST", "/api/v1/events", strings.NewReader(`{"type": "test_type", "user_id": 123}`))

	rr := httptest.NewRecorder()
	h.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
