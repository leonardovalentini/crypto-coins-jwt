package logger

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func LoggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			if r.URL.Path == "/health" {
				return
			}
			loggerWithCxt := NewLogger(r.Context())
			loggerWithCxt.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
