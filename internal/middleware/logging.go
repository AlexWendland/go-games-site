package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the status code written
// by the handler, since the standard interface doesn't expose it after the fact.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

// Logging logs the method, path, response status, and duration of each request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rr, r)

		slog.Info("api request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rr.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
