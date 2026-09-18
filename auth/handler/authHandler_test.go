package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/auth/dto"
	"github.com/leonardovalentini/crypto-coins/auth/tests/mocks/service"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var id = 1

func setup(t *testing.T) (func(), *mux.Router, *service.MockAuthService) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockAuthService(ctrl)
	ah := NewAuthHandler(mockService)

	router := mux.NewRouter()

	ah.RegisterRoutes(router)

	return func() { ctrl.Finish() }, router, mockService
}

func Test_login(t *testing.T) {
	tests := []struct {
		name            string
		username        string
		password        string
		loginError      error
		loginResponse   *http.Cookie
		cookieFound     bool
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "Login error",
			loginError:      errs.NewAuthenticationError("Invalid username or password"),
			loginResponse:   nil,
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Invalid username or password",
		},
		{
			name:       "Success",
			loginError: nil,
			loginResponse: &http.Cookie{
				Name: "session_token",
			},
			cookieFound:     true,
			expectedCode:    http.StatusOK,
			expectedMessage: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t)
			defer teardown()

			if tt.loginResponse != nil || tt.loginError != nil {
				mockService.EXPECT().Login(gomock.Any(), tt.username, tt.password).Return(tt.loginResponse, tt.loginError)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			err := writer.WriteField("username", tt.username)
			if err != nil {
				assert.Error(t, err, "Failed to write field username")
			}

			err = writer.WriteField("password", tt.password)
			if err != nil {
				assert.Error(t, err, "Failed to write field password")
			}

			err = writer.Close()
			if err != nil {
				assert.Error(t, err, "Failed to close multipart writer")
			}

			request, _ := http.NewRequest(http.MethodPost, "/auth/login", body)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			resp := recorder.Result()
			cookies := recorder.Result().Cookies()
			defer resp.Body.Close()

			var result helpers.HttpError
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				assert.Error(t, err, "Error reading body")
			}

			assert.Equal(t, tt.expectedCode, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, tt.expectedCode))
			assert.Equal(t, tt.expectedMessage, result.Message, fmt.Sprintf("Wrong response message: got %s, want %s", result.Message, tt.expectedMessage))

			var found bool
			for _, cookie := range cookies {
				if cookie.Name == "session_token" {
					found = true
				}
			}

			assert.Equal(t, tt.cookieFound, found, fmt.Sprintf("Wrong response cookie found: got %t, want %t", found, tt.cookieFound))

		})
	}
}

func Test_register_errors(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		registerError    error
		registerResponse *int
		cookieFound      bool
		expectedCode     int
		expectedMessage  string
	}{
		{
			name:             "Login error",
			registerError:    errs.NewAuthenticationError("The user is already registered."),
			registerResponse: nil,
			expectedCode:     http.StatusUnauthorized,
			expectedMessage:  "The user is already registered.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t)
			defer teardown()

			if tt.registerResponse != nil || tt.registerError != nil {
				mockService.EXPECT().Register(gomock.Any(), tt.username, tt.password).Return(tt.registerResponse, tt.registerError)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			err := writer.WriteField("username", tt.username)
			if err != nil {
				assert.Error(t, err, "Failed to write field username")
			}

			err = writer.WriteField("password", tt.password)
			if err != nil {
				assert.Error(t, err, "Failed to write field password")
			}

			err = writer.Close()
			if err != nil {
				assert.Error(t, err, "Failed to close multipart writer")
			}

			request, _ := http.NewRequest(http.MethodPost, "/auth/register", body)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			resp := recorder.Result()
			defer resp.Body.Close()

			var result helpers.HttpError
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				assert.Error(t, err, "Error reading body")
			}

			assert.Equal(t, tt.expectedCode, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, tt.expectedCode))
			assert.Equal(t, tt.expectedMessage, result.Message, fmt.Sprintf("Wrong response message: got %s, want %s", result.Message, tt.expectedMessage))

		})
	}
}
func Test_register_success(t *testing.T) {

	teardown, router, mockService := setup(t)
	defer teardown()
	id := 1
	username := "username"
	password := "password"

	mockService.EXPECT().Register(gomock.Any(), username, password).Return(&id, nil)

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	err := writer.WriteField("username", username)
	if err != nil {
		assert.Error(t, err, "Failed to write field username")
	}

	err = writer.WriteField("password", password)
	if err != nil {
		assert.Error(t, err, "Failed to write field password")
	}

	err = writer.Close()
	if err != nil {
		assert.Error(t, err, "Failed to close multipart writer")
	}

	request, _ := http.NewRequest(http.MethodPost, "/auth/register", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.ResponseCreated
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}

	assert.Equal(t, http.StatusCreated, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusCreated))
	assert.Equal(t, id, result.Id, fmt.Sprintf("Wrong response id: got %d, want %d", result.Id, id))
}

func Test_verify_errors(t *testing.T) {
	tests := []struct {
		name              string
		cookie            string
		getUserIdError    error
		getUserIdResponse *string
		expectedCode      int
		expectedMessage   string
	}{
		{
			name:            "No cookie error",
			cookie:          "no cookie",
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Session cookie not found",
		},
		{
			name:              "No session error",
			cookie:            "with cookie",
			getUserIdResponse: nil,
			getUserIdError:    errs.NewAuthenticationError("Unauthorized: No session found"),
			expectedCode:      http.StatusUnauthorized,
			expectedMessage:   "Unauthorized: No session found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t)
			defer teardown()

			cookie := &http.Cookie{
				Name:  "session_token",
				Value: "xyz123abc789",
			}

			if tt.getUserIdResponse != nil || tt.getUserIdError != nil {
				mockService.EXPECT().GetUserId(gomock.Any(), cookie).Return(tt.getUserIdResponse, tt.getUserIdError)
			}

			request, _ := http.NewRequest(http.MethodGet, "/auth/verify", nil)

			if tt.cookie == "with cookie" {
				request.AddCookie(cookie)
			}
			if tt.cookie == "malformed cookie" {
				request.Header.Add("Cookie", "session_token=\bad_value")
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			resp := recorder.Result()
			defer resp.Body.Close()

			var result helpers.HttpError
			err := json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				assert.Error(t, err, "Error reading body")
			}

			assert.Equal(t, tt.expectedCode, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, tt.expectedCode))
			assert.Equal(t, tt.expectedMessage, result.Message, fmt.Sprintf("Wrong response message: got %s, want %s", result.Message, tt.expectedMessage))

		})
	}
}

func Test_verify_success(t *testing.T) {
	teardown, router, mockService := setup(t)
	defer teardown()

	cookie := &http.Cookie{
		Name:  "session_token",
		Value: "xyz123abc789",
	}
	getUserIdResponse := "1"
	mockService.EXPECT().GetUserId(gomock.Any(), cookie).Return(&getUserIdResponse, nil)

	request, _ := http.NewRequest(http.MethodGet, "/auth/verify", nil)

	request.AddCookie(cookie)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.VerifyResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}
	expectedResponse := dto.VerifyResponse{Id: "1"}

	assert.Equal(t, http.StatusOK, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.Equal(t, expectedResponse, result, fmt.Sprintf("Wrong response message: got %v, want %v", result, expectedResponse))

}

func Test_logout_errors(t *testing.T) {
	tests := []struct {
		name            string
		cookie          string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "No cookie error",
			cookie:          "no cookie",
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Session cookie not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, _ := setup(t)
			defer teardown()

			cookie := &http.Cookie{
				Name:  "session_token",
				Value: "xyz123abc789",
			}

			request, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)

			if tt.cookie == "with cookie" {
				request.AddCookie(cookie)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			resp := recorder.Result()
			defer resp.Body.Close()

			var result helpers.HttpError
			err := json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				assert.Error(t, err, "Error reading body")
			}

			assert.Equal(t, tt.expectedCode, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, tt.expectedCode))
			assert.Equal(t, tt.expectedMessage, result.Message, fmt.Sprintf("Wrong response message: got %s, want %s", result.Message, tt.expectedMessage))

		})
	}
}

func Test_logout_success(t *testing.T) {
	teardown, router, mockService := setup(t)
	defer teardown()

	cookie := &http.Cookie{
		Name:  "session_token",
		Value: "xyz123abc789",
	}

	mockService.EXPECT().Logout(cookie).Return()

	request, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)

	request.AddCookie(cookie)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	cookies := recorder.Result().Cookies()
	defer resp.Body.Close()

	var result dto.Response
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}
	expectedResponse := dto.Response{Message: "OK"}

	assert.Equal(t, http.StatusOK, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.Equal(t, expectedResponse, result, fmt.Sprintf("Wrong response message: got %v, want %v", result, expectedResponse))

	var respCookie *http.Cookie

	for _, cookie := range cookies {
		if cookie.Name == "session_token" {
			respCookie = cookie
		}
	}
	// expectedCookie := &http.Cookie{
	// 	Name:        "session_token",
	// 	Value:       "",
	// 	Quoted:      false,
	// 	Path:        "/",
	// 	Domain:      "",
	// 	Expires:     time.Date(2026, time.August, 31, 12, 58, 39, 0, time.UTC),
	// 	RawExpires:  "Mon, 31 Aug 2026 12:58:39 GMT",
	// 	MaxAge:      0,
	// 	Secure:      false,
	// 	HttpOnly:    true,
	// 	SameSite:    0,
	// 	Partitioned: false,
	// 	Raw:         "session_token=; Path=/; Expires=Mon, 31 Aug 2026 12:58:39 GMT; HttpOnly",
	// 	Unparsed:    []string(nil),
	// }
	assert.True(t, respCookie.Expires.Before(time.Now()), "Cookie should be expired")

}
