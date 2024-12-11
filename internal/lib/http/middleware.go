package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"unicode"

	"github.com/go-chi/jwtauth"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
)

type contextKey struct {
	Key string
}

var (
	CtxKeyLogger = &contextKey{"logger"}
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

			log.Info("request handled", slog.Duration("elapsed", time.Duration(time.Since(t).Milliseconds())))
		})
	}
}

func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			msg := []rune(err.Error())
			msg[0] = unicode.ToUpper(msg[0])
			ErrUnauthorized(w, r, string(msg))
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			ErrUnauthorized(w, r, "Token is unauthorized")
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}
