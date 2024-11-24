package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	CtxKeyLogger = "logger"
)

var (
	ErrLoggerNotFound = errors.New("logger not found in context")
)

func GetCtxLogger(ctx context.Context) *slog.Logger {
	return ctx.Value(CtxKeyLogger).(*slog.Logger)
}

// TraceID middleware adds traceID to each request if it is not set
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		r.Header.Set("X-Trace-Id", traceID)

		next.ServeHTTP(w, r)
	})
}

// Logging middleware for logging requests
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger

			traceID := r.Header.Get("X-Trace-ID")
			if traceID != "" {
				log = log.With(slog.String("trace_id", traceID))
			}

			ctx := context.WithValue(r.Context(), CtxKeyLogger, log)

			log.Info("incoming request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			t := time.Now()

			next.ServeHTTP(w, r.WithContext(ctx))

			log.Info("request handled", slog.Duration("elapsed", time.Since(t)))
		})
	}
}
