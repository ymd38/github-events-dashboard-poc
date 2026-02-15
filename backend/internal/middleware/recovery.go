package middleware

import (
	"net/http"
	"runtime/debug"
)

// Recovery is an HTTP middleware that recovers from panics and returns 500.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				LogEvent("error", "panic recovered", map[string]interface{}{
					"error": err,
					"stack": string(debug.Stack()),
					"path":  r.URL.Path,
				})
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
