package server

import (
    "fmt"
	"net/http"
	"github.com/golang-jwt/jwt"
	//"time"
)

var sampleSecretKey = []byte("123")

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

func AuthenticationJwt(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
	    if r.Header["Api-Token"] == nil {
	        w.Write([]byte("Can not find token in header"));
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        token, _ := jwt.Parse(r.Header["Api-Token"][0], func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("There was an error in parsing")
            }
            return sampleSecretKey, nil
        })

        if token == nil {
            w.Write([]byte("Invalid token"));
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        if !token.Valid {
            w.Write([]byte("Not valid token"));
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)

        if !ok {
            w.Write([]byte("couldn't parse claims"));
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        if claims["user_id"] == nil {
            w.Write([]byte("user_id not found"));
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
