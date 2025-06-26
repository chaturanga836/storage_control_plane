// internal/logger/logging.go
package logger

import (
	"log"
	"net/http"
)

// LogRequest is a middleware that logs incoming HTTP requests.
// Use this in main.go or router setup.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📥 %s %s", r.Method, r.URL.Path)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
