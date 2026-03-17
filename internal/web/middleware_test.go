package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// nextOK is a trivial handler that writes 200 OK.
var nextOK = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestCors_NoOriginHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_WildcardByDefault(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
	assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Headers"))
	assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Methods"))
}

func TestCors_AllowedOriginMatches(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "https://app.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_AllowedOriginWithWhitespace(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", " https://app.example.com , https://admin.example.com ")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://admin.example.com")
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, "https://admin.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_OriginNotAllowed(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_OptionsPreflightReturns204(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestCors_OptionsWithoutOriginReturns204(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()
	Cors(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestCors_PassthroughToNextHandler(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	Cors(next).ServeHTTP(httptest.NewRecorder(), req)

	assert.True(t, called)
}

func TestCors_OptionsDoesNotCallNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	Cors(next).ServeHTTP(httptest.NewRecorder(), req)

	assert.False(t, called)
}

// --- Auth middleware ---

const testToken = "test-bearer-token-abc123"

func TestAuth_ValidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	rr := httptest.NewRecorder()

	Auth(testToken)(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()

	Auth(testToken)(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	Auth(testToken)(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuth_BearerPrefixStripped(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer  "+testToken)
	rr := httptest.NewRecorder()

	Auth(testToken)(nextOK).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuth_PassthroughToNextHandler(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	Auth(testToken)(next).ServeHTTP(httptest.NewRecorder(), req)

	assert.True(t, called)
}

func TestAuth_DoesNotCallNextOnInvalid(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad")
	Auth(testToken)(next).ServeHTTP(httptest.NewRecorder(), req)

	assert.False(t, called)
}
