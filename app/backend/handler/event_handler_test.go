package handler
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"io"
	//"errors"
    //"database/sql"
	repository "micro-rest-events/v1/app/backend/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mock_event "micro-rest-events/v1/app/backend/repository/mocks"
	"github.com/go-chi/chi/v5"
	"strings"
	"fmt"
)

func TestOnCreateEvent(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)

	handler := Handler{
		EventRepository: mockRepo,
	}

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
	handler.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnCreateBatchEvent(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)

	handler := Handler{
		EventRepository: mockRepo,
	}

    payload := map[string]interface{}{
        "type":    "test_type",
        "users": []string{
            "123-1234-11111",
            "123-1234-22222",
        },
        "caption": "test_caption",
        "body":    "test_body",
    }

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnCreateBatchEvent_TypeNotFound(t *testing.T) {
	handler := Handler{
	}

    payload := map[string]interface{}{
        "users": []string{
            "123-1234-11111",
            "123-1234-22222",
        },
        "caption": "test_caption",
        "body":    "test_body",
    }

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateBatchEvent_BadData(t *testing.T) {
	handler := Handler{
	}

	req, err := http.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader([]byte("bla")))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.OnCreateBatchEvents(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_EmptyPayload(t *testing.T) {
	handler := Handler{}

	payload := map[string]interface{}{
		"type":    "",
		"user_id": nil,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_BadJsonRequest(t *testing.T) {
	handler := Handler{}

	req, err := http.NewRequest("POST", "/api/v1/events", bytes.NewReader([]byte("Bad json request")))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnCreateEvent_RepositoryError(t *testing.T) {
	mockRepo := new(mock_event.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(fmt.Errorf("Unexpected error"))

	handler := Handler{
		EventRepository: mockRepo,
	}

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
	handler.OnCreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestOnGetEventsByUserId(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockEventOne := repository.Event{
        Uuid:   "test_uuid",
        UserId: "123",
        Type:   "test_type",
        Status: "new",
    }

    mockEvent := []repository.Event{mockEventOne}

    mockRepo.On("GetAllByUserId", "123", mock.AnythingOfType("repository.Query")).Return(mockEvent, nil)

    h := Handler{
        EventRepository: mockRepo,
    }

     r := chi.NewRouter()
    r.Get("/api/v1/events/users/{id}", h.OnGetEventsByUserId)

    req, err := http.NewRequest("GET", "/api/v1/events/users/123", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnGetEventsByUserId_NotFound(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    var mockEvent []repository.Event
    mockRepo.On("GetAllByUserId", "123", mock.AnythingOfType("repository.Query")).Return(mockEvent, fmt.Errorf("Event not found"))

    h := Handler{
        EventRepository: mockRepo,
    }

     r := chi.NewRouter()
    r.Get("/api/v1/events/users/{id}", h.OnGetEventsByUserId)

    req, err := http.NewRequest("GET", "/api/v1/events/users/123", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    mockRepo.AssertExpectations(t)
}

// func TestOnGetOneEventByUserId(t *testing.T) {
//     mockRepo := new(mock_event.MockEventRepository)
//     mockEvent := repository.Event{
//         Uuid:   "test_uuid",
//         UserId: "123",
//         Type:   "test_type",
//         Status: "new",
//     }
//     mockRepo.On("GetOneByUserId", "123").Return(mockEvent, nil)
//
//     h := Handler{
//         EventRepository: mockRepo,
//     }
//
//      r := chi.NewRouter()
//     r.Get("/api/v1/events/users/{id}", h.OnGetOneEventByUserId)
//
//     req, err := http.NewRequest("GET", "/api/v1/events/users/123", nil)
//     if err != nil {
//         t.Fatal(err)
//     }
//
//     rr := httptest.NewRecorder()
//     r.ServeHTTP(rr, req)
//
//     assert.Equal(t, http.StatusOK, rr.Code)
//
//     mockRepo.AssertExpectations(t)
// }
// func TestOnGetEventsByUserId_NotFound(t *testing.T) {
//     mockRepo := new(mock_event.MockEventRepository)
//     mockEvent := repository.Event{}
//     mockRepo.On("GetByUserId", "123").Return(mockEvent, fmt.Errorf("Event not found"))
//
//     h := Handler{
//         EventRepository: mockRepo,
//     }
//
//      r := chi.NewRouter()
//     r.Get("/api/v1/events/users/{id}", h.OnGetEventsByUserId)
//
//     req, err := http.NewRequest("GET", "/api/v1/events/users/123", nil)
//     if err != nil {
//         t.Fatal(err)
//     }
//
//     rr := httptest.NewRecorder()
//     r.ServeHTTP(rr, req)
//
//     assert.Equal(t, http.StatusOK, rr.Code)
//
//     mockRepo.AssertExpectations(t)
// }

func TestOnChangeEvent(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeStatus", "test_uuid", mock.AnythingOfType("repository.Event")).Return(int64(1), nil)

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

    payload := `{"status": "new_status", "message": "new_message"}`
    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnChangeBatchEvents(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeStatus", "123-1234-11111", mock.AnythingOfType("repository.Event")).Return(int64(1), nil)

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

    payload := map[string]interface{}{
        "status": "delivered",
        "uuids": []string{
            "123-1234-11111",
        },
    }

    payloadBytes, _ := json.Marshal(payload)

    req, err := http.NewRequest("POST", "/api/v1/events/change/batch", bytes.NewReader(payloadBytes))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnChangeBatchEvents_NotFoundUuids(t *testing.T) {
    r := chi.NewRouter()
    h := Handler{}
    r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

    payload := map[string]interface{}{
        "status": "delivered",
    }

    payloadBytes, _ := json.Marshal(payload)

    req, err := http.NewRequest("POST", "/api/v1/events/change/batch", bytes.NewReader(payloadBytes))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeBatchEvents_BadJson(t *testing.T) {
    r := chi.NewRouter()
    h := Handler{}
    r.Post("/api/v1/events/change/batch", h.OnChangeBatchEvents)

    payload := `Bad Json`

    req, err := http.NewRequest("POST", "/api/v1/events/change/batch", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeEvent_BadPayload(t *testing.T) {
    r := chi.NewRouter()
    h := Handler{}
    r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader("Bad payload"))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnChangeEvent_StatusNotFound(t *testing.T) {
    r := chi.NewRouter()
    h := Handler{}
    r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

    payload := `{"message": "new_message"}`
    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }

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

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

    payload := `{"status": "new_status", "message": "new_message"}`
    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusNotFound, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnChangeEvent_RepositoryReturnError(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeStatus", "test_uuid", mock.AnythingOfType("repository.Event")).Return(int64(0), fmt.Errorf("Event not found"))

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}", h.OnChangeEvent)

    payload := `{"status": "new_status", "message": "new_message"}`
    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid", strings.NewReader(payload))
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnSetSeen(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(1), nil)

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnSetSeenNotFound(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(0), nil)

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusNotFound, rr.Code)

    mockRepo.AssertExpectations(t)
}

func TestOnSetSeenError(t *testing.T) {
    mockRepo := new(mock_event.MockEventRepository)
    mockRepo.On("ChangeIsSeen", "test_uuid").Return(int64(0), fmt.Errorf("Some error"))

    r := chi.NewRouter()
    h := Handler{
        EventRepository: mockRepo,
    }
    r.Post("/api/v1/events/{uuid}/seen", h.OnSetSeen)

    req, err := http.NewRequest("POST", "/api/v1/events/test_uuid/seen", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)

    mockRepo.AssertExpectations(t)
}
