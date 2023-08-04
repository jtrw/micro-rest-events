package server

import (
	"net/http"
)


func Authentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
	    token := r.Header.Get("Api-Token")
	    if token != "111" {
	        w.Write([]byte("Unauthorized"));
	        w.WriteHeader(http.StatusUnauthorized)
        	return

	    }
	    next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func Cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*") // change this later
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}
