package handler

import (
    "net/http"
    "testing"
    "net/http/httptest"
    "micro-rest-events/v1/app/backend/repository"
    "github.com/stretchr/testify/assert"
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
