package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

var jsonLogger = log.New(os.Stdout, "", 0)

// LogEntry represents a structured log entry in JSON format.
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Duration   string `json:"duration"`
	RemoteAddr string `json:"remote_addr"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger is an HTTP middleware that logs each request in structured JSON format.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		entry := LogEntry{
			Timestamp:  start.UTC().Format(time.RFC3339),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: rw.statusCode,
			Duration:   time.Since(start).String(),
			RemoteAddr: r.RemoteAddr,
		}
		data, _ := json.Marshal(entry)
		jsonLogger.Println(string(data))
	})
}

// LogEvent logs an application-level event in structured JSON format.
func LogEvent(level string, message string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}
	for k, v := range fields {
		entry[k] = v
	}
	data, _ := json.Marshal(entry)
	jsonLogger.Println(string(data))
}
