// Package middleware provides HTTP middleware functions for request handling, logging, and security.
package middleware

import "net/http"

// SecurityHeaders adds HTTP response headers to improve security,
// including Strict-Transport-Security (HSTS), X-Content-Type-Options,
// and X-Frame-Options as recommended for secure web applications.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enforce HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Enable XSS filtering in the browser
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}
