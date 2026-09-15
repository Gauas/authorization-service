package forwardauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			r.Header.Set("X-Request-ID", requestID())
		}
		started := time.Now()
		response := &responseWriter{ResponseWriter: w}
		defer func() {
			failure := recover()
			if failure != nil {
				slog.Error("http.panic", "request_id", r.Header.Get("X-Request-ID"), "error", fmt.Sprint(failure))
				http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			if r.URL.Path == "/health/live" || r.URL.Path == "/metrics" {
				return
			}
			slog.Info("http.request",
				"request_id", r.Header.Get("X-Request-ID"),
				"method", r.Method,
				"path", r.URL.Path,
				"status", response.status,
				"bytes", response.bytes,
				"latency_ms", time.Since(started).Milliseconds(),
			)
		}()
		next.ServeHTTP(response, r)
	})
}

func requestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}
