package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	sessionName            = "github-dashboard-session"
	sessionUser            = "user_id"
)

// Auth returns an HTTP middleware that validates the user session.
// Unauthenticated requests receive a 401 response.
func Auth(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, sessionName)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid session"})
				return
			}
			val, ok := session.Values[sessionUser]
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
				return
			}
			userID, ok := val.(int64)
			if !ok || userID == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid session data"})
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) int64 {
	val, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0
	}
	return val
}
