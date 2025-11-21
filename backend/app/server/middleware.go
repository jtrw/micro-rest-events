package server

import (
	"net/http"
	"os"
	"strings"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Get allowed origins from environment variable, default to wildcard for backward compatibility
			allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
			var allowedOrigins []string

			if allowedOriginsEnv != "" {
				allowedOrigins = strings.Split(allowedOriginsEnv, ",")
			} else {
				// Default: allow all origins (for backward compatibility)
				// IMPORTANT: In production, set ALLOWED_ORIGINS environment variable
				allowedOrigins = []string{"*"}
			}

			// Check if origin is allowed
			isAllowed := false
			for _, allowed := range allowedOrigins {
				allowed = strings.TrimSpace(allowed)
				if allowed == "*" || allowed == origin {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				if allowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Api-Token")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
