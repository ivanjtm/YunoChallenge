package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// Chain applies middlewares in reverse order so that the first middleware
// in the list is the outermost (executed first on each request).
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs each request with method, path, status code, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %s",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			rec.statusCode,
			duration,
		)
	})
}

// ContentTypeMiddleware ensures that POST, PUT, and PATCH requests contain
// an application/json Content-Type header.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			ct := r.Header.Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
					"Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics, logs the stack trace,
// and returns a 500 Internal Server Error.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				log.Printf("[PANIC] %v\n%s", rv, debug.Stack())
				WriteError(w, http.StatusInternalServerError, "internal_error",
					"An unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// WriteJSON writes a JSON response with the given HTTP status code and data.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// WriteError writes a JSON error response with the given status, error code, and message.
func WriteError(w http.ResponseWriter, status int, errCode string, message string) {
	WriteJSON(w, status, errorResponse{
		Error:   errCode,
		Message: message,
	})
}
