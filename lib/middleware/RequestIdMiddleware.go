package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			w.Header().Set("X-Request-ID", requestID)
		}

		ctx := contextKey.WithRequestId(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
