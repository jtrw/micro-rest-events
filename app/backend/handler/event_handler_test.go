package handler

import (
    "net/http"
    "testing"
    "net/http/httptest"
    "micro-rest-events/v1/app/backend/repository"
    "github.com/stretchr/testify/assert"
    eventMock "micro-rest-events/v1/app/backend/repository/mocks"
    event "micro-rest-events/v1/app/backend/repository"
    "encoding/json"
    "bytes"
    //"github.com/stretchr/testify/require"
)
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

func TestOnCreateEvent_Success(t *testing.T) {
    t.Setenv("POSTGRES_DSN", "host=localhost port=5432 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable")
    repo := repository.ConnectDB()
    handler := NewHandler(repo.Connection)
	// Підготовка
	//handler := Handler{} // Замініть на створення вашого об'єкта Handler
	mockRepo := &eventMock.MockEventRepository{} // Замініть на вашу моковану реалізацію репозиторію

	expectedUserID := 123
	expectedEvent := event.Event{ /* ініціалізуйте очікувану подію тут */ }
	mockRepo.On("Create", expectedEvent).Return(nil)

	requestData := JSON{"type": "new", "user_id": expectedUserID}
	reqBody, _ := json.Marshal(requestData)
	req := httptest.NewRequest("POST", "/events", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.OnCreateEvent(rec, req)

	// Перевірка
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	// Перевірте тіло відповіді тут, наприклад, через розпакування JSON

	mockRepo.AssertExpectations(t)
}
