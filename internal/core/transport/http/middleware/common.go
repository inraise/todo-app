package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/inraise/todo-app/internal/core/logger"
	"go.uber.org/zap"
)

const (
	reqIDHeader = "X-Request-ID"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(reqIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()

				r.Header.Set(reqIDHeader, requestID)
				w.Header().Set(reqIDHeader, requestID)

				next.ServeHTTP(w, r)
			}
		})
	}
}

func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(reqIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)

			ctx := context.WithValue(r.Context(), "log", l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
