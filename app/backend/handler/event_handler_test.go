package handler
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	//"errors"
    //"database/sql"
	repository "micro-rest-events/v1/app/backend/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mock_repository "micro-rest-events/v1/app/backend/repository/mocks"
	"github.com/go-chi/chi/v5"
	"strings"
	"fmt"
)

func TestOnCreateEvent(t *testing.T) {
	mockRepo := new(mock_repository.MockEventRepository)
	mockRepo.On("Create", mock.AnythingOfType("repository.Event")).Return(nil)

	handler := Handler{
		EventRepository: mockRepo,
	}

	payload := map[string]interface{}{
		"type":    "test_type",
		"user_id": 123,
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


func TestOnGetEventsByUserIdNotFound(t *testing.T) {
    //t.Setenv("POSTGRES_DSN", "host=localhost port=5432 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable")
    //repo := repository.ConnectDB()
    //h := NewHandler(repo.Connection)

    mockRepo := new(mock_repository.MockEventRepository)
    mockEvent := repository.Event{
        Uuid:   "test_uuid",
        UserId: 123,
        Type:   "test_type",
        Status: "new",
    }
    mockRepo.On("GetByUserId", 123).Return(mockEvent, nil)

    h := Handler{
        EventRepository: mockRepo,
    }

     r := chi.NewRouter()
    r.Get("/get-events/{id}", h.OnGetEventsByUserId)

    req, err := http.NewRequest("GET", "/get-events/123", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOnChangeEvent(t *testing.T) {
    mockRepo := new(mock_repository.MockEventRepository)
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

func TestOnSetSeen(t *testing.T) {
    mockRepo := new(mock_repository.MockEventRepository)
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
    mockRepo := new(mock_repository.MockEventRepository)
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
    mockRepo := new(mock_repository.MockEventRepository)
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
