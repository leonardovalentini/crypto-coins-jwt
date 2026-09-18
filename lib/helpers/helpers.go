package helpers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type HttpError struct {
	Message string `json:"message"`
}

const fallbackJSONError = `{"message":"Internal Server Error"}`
const fallbackXMLrror = `<error><message>Internal server error during XML encoding</message></error>`

func WriteResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	loggerWithCxt := logger.NewLogger(r.Context())
	if r.Header.Get("Accept") == "application/xml" {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(statusCode)
		if err := xml.NewEncoder(w).Encode(data); err != nil {
			loggerWithCxt.Error("Failed to encode response: " + err.Error())
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusInternalServerError)
			if _, writeErr := w.Write([]byte(fallbackXMLrror)); writeErr != nil {
				loggerWithCxt.Error("Failed to write fallback XML to client: " + writeErr.Error())
			}
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		loggerWithCxt.Error("Failed to encode response: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write([]byte(fallbackJSONError)); writeErr != nil {
			loggerWithCxt.Error("Failed to write fallback JSON to client: " + writeErr.Error())
		}
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err, ok := errors.AsType[*errs.AppError](err); ok {
		WriteResponse(w, r, err.Code, HttpError{Message: err.Message})
		return
	}

	WriteResponse(w, r, http.StatusInternalServerError, HttpError{Message: "Internal Server Error"})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				loggerWithCxt := logger.NewLogger(r.Context())
				loggerWithCxt.Error(fmt.Sprintf("Encountered unexpected error: %v", err))
				newErr := errs.NewUnexpectedError("Internal Server Error")
				HandleError(w, r, newErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func GetAttribute(object map[string]any, attribute string) (map[string]any, error) {
	levels := strings.Split(attribute, ".")

	for _, level := range levels {
		attr, ok := object[level]
		if !ok {
			return nil, errs.NewUnexpectedError("Error decoding response")
		}
		object, ok = attr.(map[string]any)
		if !ok {
			return nil, errs.NewUnexpectedError("Error decoding response")
		}

	}

	return object, nil
}

func BuildUrl(baseURL string, queryParams map[string]string) string {
	params := url.Values{}

	for key, value := range queryParams {
		params.Set(key, value)
	}

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}
