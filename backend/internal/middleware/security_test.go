package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "Valid requests receive standard security headers",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-Content-Type-Options":    "nosniff",
				"X-Frame-Options":           "DENY",
				"X-XSS-Protection":          "1; mode=block",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := SecurityHeaders(nextHandler)

			req := httptest.NewRequest(tt.method, "http://example.com/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, res.StatusCode)
			}

			for key, expectedValue := range tt.expectedHeaders {
				actualValue := res.Header.Get(key)
				if actualValue != expectedValue {
					t.Errorf("expected header %q to be %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}
