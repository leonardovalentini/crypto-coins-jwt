package authMiddleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type TargetResponse struct {
	Id string `json:"id"`
}

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loggerWithCxt := logger.NewLogger(r.Context())
			cookie, err := r.Cookie("session_token")
			if err != nil {
				if err == http.ErrNoCookie {
					loggerWithCxt.Error("Unauthorized: No session token found: " + err.Error())
					helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: No session token found"))
					return
				}
				helpers.HandleError(w, r, errs.NewAuthenticationError("Invalid session"))
				return
			}

			address := os.Getenv("AUTH_HOST")
			port := os.Getenv("AUTH_PORT")
			url := fmt.Sprintf("http://%s:%s/auth/verify", address, port)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				loggerWithCxt.Error("Unauthorized: Unexpected error while validating credentials. " + err.Error())
				helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: Unexpected error while validating credentials."))
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				loggerWithCxt.Error("Unauthorized: Error verifying credentials. " + err.Error())
				helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: Error verifying credentials."))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				loggerWithCxt.Error(fmt.Sprintf("Unauthorized: Error verifying credentials. Expected status 200, got: %d", resp.StatusCode))
				helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: Error verifying credentials."))
				return
			}

			var result TargetResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				loggerWithCxt.Error("Unauthorized: Error verifying credentials. " + err.Error())
				helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: Error verifying credentials."))
				return
			}

			userId := result.Id
			if err != nil {
				loggerWithCxt.Error("Unauthorized: " + err.Error())
				helpers.HandleError(w, r, errs.NewAuthenticationError("Unauthorized: Error verifying credentials."))
				return
			}
			ctx := *contextKey.SetUserId(r, userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
