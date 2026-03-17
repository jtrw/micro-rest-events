package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookie   = "session"
	sessionDuration = 24 * time.Hour
)

// AuthMiddleware redirects unauthenticated requests to /login.
// Public paths (/login, /static/, /ping, /robots.txt, /api/) bypass the check.
// Auth is skipped entirely when AuthPassword is empty.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.AuthPassword == "" {
			next.ServeHTTP(w, r)
			return
		}
		p := r.URL.Path
		if p == "/login" || p == "/logout" ||
			strings.HasPrefix(p, "/static/") ||
			strings.HasPrefix(p, "/api/") ||
			p == "/ping" || p == "/robots.txt" {
			next.ServeHTTP(w, r)
			return
		}
		c, err := r.Cookie(sessionCookie)
		if err != nil || !s.verifySession(c.Value) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if s.AuthPassword == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := s.tmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": ""}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	okUser := hmac.Equal([]byte(username), []byte(s.AuthLogin))
	okPass := hmac.Equal([]byte(password), []byte(s.AuthPassword))

	if !okUser || !okPass {
		slog.Warn("failed login attempt", "username", username, "ip", r.RemoteAddr)
		if err := s.tmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "invalid username or password"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := s.createSession(username)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// createSession returns a signed token: "username:expires:hmac".
func (s *Server) createSession(username string) string {
	expires := time.Now().Add(sessionDuration).Unix()
	payload := fmt.Sprintf("%s:%d", username, expires)
	return payload + ":" + s.sign(payload)
}

// verifySession validates signature and expiry of a session token.
func (s *Server) verifySession(value string) bool {
	idx := strings.LastIndex(value, ":")
	if idx < 0 {
		return false
	}
	payload, sig := value[:idx], value[idx+1:]
	if !hmac.Equal([]byte(sig), []byte(s.sign(payload))) {
		return false
	}
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		return false
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	return true
}

func (s *Server) sign(payload string) string {
	key := []byte("session:" + s.Secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
