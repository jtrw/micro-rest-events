package handler

import (
    "net/http"
    "testing"
    "net/http/httptest"
    "micro-rest-events/v1/app/backend/repository"
    "github.com/stretchr/testify/assert"
    "strings"
    "github.com/golang/mock/gomock"
   // eventMock "micro-rest-events/v1/app/backend/repository/mocks"
     mock_repository "micro-rest-events/v1/app/backend/repository/mocks"
   // event "micro-rest-events/v1/app/backend/repository"
 //   "encoding/json"
 //   "bytes"
  //  "github.com/DATA-DOG/go-sqlmock"
    //"github.com/stretchr/testify/require"
)

//const STATUS_NEW = "new"

//type JSON map[string]interface{}

// type Handler struct {
// 	Connection *sql.DB
// 	EventRepository *event.EventRepository
// }
//
// func NewHandler(conn *sql.DB) Handler {
// 	return Handler{Connection: conn, EventRepository: event.NewEventRepository(conn)}
// }

// func (h Handler) NewEventRepository() *event.EventRepository {
//     return event.NewEventRepository(h.Connection)
// }
func TestOnGetEventsByUserIdNotFound(t *testing.T) {
    t.Setenv("POSTGRES_DSN", "host=localhost port=5432 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable")
    repo := repository.ConnectDB()
    h := NewHandler(repo.Connection)
    // Create a request to pass to our handler. We don't have any query parameters for now, so we'll
    // pass 'nil' as the third parameter.
    req, err := http.NewRequest("GET", "/events/users/9999999", nil)
    if err != nil {
        t.Fatal(err)
    }
    // We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(h.OnGetEventsByUserId)
    // Our handlers satisfy http.Handler, so we can call their ServeHTTP method
    // directly and pass in our Request and ResponseRecorder.
    handler.ServeHTTP(rr, req)

     assert.NoError(t, err)
     assert.Equal(t, 404, rr.Code)

     //assert.Equal(t, "{\"status\": \"not_found\", \"message\": \"Not Found\"}", rr.Body.String())
}

func TestOnCreateEvent(t *testing.T) {
	// Arrange
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEventRepository := mock_repository.NewMockEventRepository(mockCtrl)
	handler := Handler{Connection: nil, EventRepository: mockEventRepository}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"type":"test", "user_id":1}`))

	// Act
	mockEventRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	handler.OnCreateEvent(w, r)

	// Assert
	status := w.Code
	if status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}

	body := w.Body.String()
	expected := `{"status":"ok","uuid":"<uuid>"}`
	if body != expected {
		t.Errorf("Expected body %s, got %s", expected, body)
	}
}

// func TestOnCreateEvent_Success(t *testing.T) {
//     db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer db.Close()
//
// 	//mock.ExpectBegin()
//     mock.ExpectExec(`INSERT INTO "events"`).WithArgs(1, 2, 3, 4).WillReturnResult(sqlmock.NewResult(1, 1))
//     //mock.ExpectCommit()
//
//     //t.Setenv("POSTGRES_DSN", "host=localhost port=5432 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable")
//     //repo := repository.ConnectDB()
//     //handler := NewHandler(db)
//
// 	mockRepo := &eventMock.MockEventRepository{}
//
// 	handler := Handler{Connection: db, EventRepository: mockRepo}
//
// 	expectedUserID := 123
// 	expectedEvent := event.Event{}
// 	mockRepo.On("Create", expectedEvent).Return(nil)
//
// 	requestData := JSON{"type": "new", "user_id": expectedUserID}
// 	reqBody, _ := json.Marshal(requestData)
// 	req := httptest.NewRequest("POST", "/events", bytes.NewReader(reqBody))
// 	rec := httptest.NewRecorder()
//
// 	handler.OnCreateEvent(rec, req)
//
// 	assert.Equal(t, http.StatusCreated, rec.Code)
// 	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
//
// 	//mockRepo.AssertExpectations(t)
// }
