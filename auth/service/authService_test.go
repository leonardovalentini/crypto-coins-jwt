package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/leonardovalentini/crypto-coins/auth/domain"
	domainMock "github.com/leonardovalentini/crypto-coins/auth/tests/mocks/domain"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var getCurrentTime = func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }
var randomGeneratorResponseError int = 0
var randomGeneratorResponseSuccess int = 1
var exist = true
var doNotExist = false
var id = 1
var sid = "1"

func mockRandomGenerator(randomGeneratorResponse *int, randomGeneratorError error) func(b []byte) (n int, err error) {
	if randomGeneratorResponse != nil {
		return func(b []byte) (n int, err error) {
			return *randomGeneratorResponse, randomGeneratorError
		}
	} else {
		return rand.Read
	}
}

func setup(t *testing.T, randomGeneratorResponse *int, randomGeneratorError error) (func(), AuthService, *domainMock.MockUserRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := domainMock.NewMockUserRepository(ctrl)
	mockService := NewAuthService(mockRepo, mockRandomGenerator(randomGeneratorResponse, randomGeneratorError), getCurrentTime)
	return func() {
		mockService = nil
		defer ctrl.Finish()
	}, mockService, mockRepo
}

func Test_login(t *testing.T) {
	tests := []struct {
		name                    string
		findByUsernameError     error
		findByUsernameResponse  *domain.User
		randomGeneratorError    error
		randomGeneratorResponse *int
		expectedError           error
		expectedResponse        *http.Cookie
	}{
		{
			name:                   "Find all by unexpected error",
			findByUsernameResponse: nil,
			findByUsernameError:    errs.NewUnexpectedError("unexpected error"),
			expectedError:          errs.NewUnexpectedError("unexpected error"),
			expectedResponse:       nil,
		},
		{
			name: "bcrypt error",
			findByUsernameResponse: &domain.User{
				Id:       1,
				Username: "mock username",
				Password: "invalid-hash-format",
			},
			findByUsernameError: nil,
			expectedError:       errs.NewAuthenticationError("Invalid username or password"),
			expectedResponse:    nil,
		},
		{
			name: "Random generatorError error",
			findByUsernameResponse: &domain.User{
				Id:       1,
				Username: "mock username",
				Password: "$2a$10$XMZHoMpCG1o9toxIwCclJeupL11xbilkPikSLFJJmIZ0zntIFGKNm91",
			},
			findByUsernameError:     nil,
			randomGeneratorError:    errors.New("Unexpected error"),
			randomGeneratorResponse: &randomGeneratorResponseError,
			expectedError:           errs.NewUnexpectedError("Unexpected error"),
			expectedResponse:        nil,
		},
		{
			name: "Success",
			findByUsernameResponse: &domain.User{
				Id:       1,
				Username: "mock username",
				Password: "$2a$10$XMZHoMpCG1o9toxIwCclJeupL11xbilkPikSLFJJmIZ0zntIFGKNm91",
			},
			findByUsernameError:     nil,
			randomGeneratorResponse: &randomGeneratorResponseSuccess,
			expectedError:           nil,
			expectedResponse: &http.Cookie{
				Name:     "session_token",
				Value:    "0000000000000000000000000000000000000000000000000000000000000000",
				Path:     "/",
				Expires:  getCurrentTime().Add(30 * time.Minute),
				MaxAge:   1800,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t, tt.randomGeneratorResponse, tt.randomGeneratorError)
			defer teardown()

			username := "mock username"
			password := "secret123"

			if tt.findByUsernameResponse != nil || tt.findByUsernameError != nil {
				mockRepo.EXPECT().FindByUsername(ctx, username).Return(tt.findByUsernameResponse, tt.findByUsernameError)
			}

			resp, err := mockService.Login(ctx, username, password)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedError))
		})
	}
}

func Test_register(t *testing.T) {
	tests := []struct {
		name                          string
		checkIfUsernameExistsError    error
		checkIfUsernameExistsResponse *bool
		password                      string
		saveError                     error
		saveResponse                  *domain.User
		expectedError                 error
		expectedResponse              *int
	}{
		{
			name:                          "Check if username exists unexpected error",
			checkIfUsernameExistsResponse: nil,
			checkIfUsernameExistsError:    errs.NewUnexpectedError("unexpected error"),
			expectedError:                 errs.NewUnexpectedError("unexpected error"),
			expectedResponse:              nil,
		},
		{
			name:                          "User already exists error",
			checkIfUsernameExistsResponse: &exist,
			checkIfUsernameExistsError:    nil,
			expectedError:                 errs.NewConflictError("The user is already registered."),
			expectedResponse:              nil,
		},
		{
			name:                          "Generate from password error",
			checkIfUsernameExistsResponse: &doNotExist,
			checkIfUsernameExistsError:    nil,
			password:                      strings.Repeat("a", 73),
			expectedError:                 errs.NewUnexpectedError("An error occurred while encrypting the password."),
			expectedResponse:              nil,
		},
		{
			name:                          "Save error",
			checkIfUsernameExistsResponse: &doNotExist,
			checkIfUsernameExistsError:    nil,
			password:                      "secret123",
			saveError:                     errs.NewUnexpectedError("unexpected error"),
			saveResponse:                  nil,
			expectedError:                 errs.NewUnexpectedError("unexpected error"),
			expectedResponse:              nil,
		},
		{
			name:                          "Success",
			checkIfUsernameExistsResponse: &doNotExist,
			checkIfUsernameExistsError:    nil,
			password:                      "secret123",
			saveError:                     nil,
			saveResponse:                  &domain.User{Id: id},
			expectedError:                 nil,
			expectedResponse:              &id,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t, nil, nil)
			defer teardown()

			username := "mock username"

			if tt.checkIfUsernameExistsResponse != nil || tt.checkIfUsernameExistsError != nil {
				mockRepo.EXPECT().CheckIfUsernameExists(ctx, username).Return(tt.checkIfUsernameExistsResponse, tt.checkIfUsernameExistsError)
			}

			if tt.saveResponse != nil || tt.saveError != nil {

				mockRepo.EXPECT().Save(ctx, gomock.WantFormatter(gomock.Eq("is valid user"), gomock.Cond(func(x any) bool {
					user, ok := x.(domain.User)
					return ok && user.Username == username
				}))).Return(tt.saveResponse, tt.saveError)
			}

			resp, err := mockService.Register(ctx, username, tt.password)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedError))
		})
	}
}

func Test_logout(t *testing.T) {
	teardown, mockService, _ := setup(t, nil, nil)
	defer teardown()
	userId := "1"
	sessionToken := "mock token"

	cookie := http.Cookie{Value: sessionToken}
	// save session
	sessionMutex.Lock()
	sessionStore[sessionToken] = userId
	sessionMutex.Unlock()
	// check session exists
	sessionMutex.RLock()
	savedUserId, exists := sessionStore[cookie.Value]
	sessionMutex.RUnlock()
	assert.Equal(t, true, exists, fmt.Sprintf("Wrong exists: got %v, want %v", exists, true))
	assert.Equal(t, userId, savedUserId, fmt.Sprintf("Wrong userId: got %v, want %v", savedUserId, userId))
	// logout
	mockService.Logout(&cookie)
	// check session exists
	sessionMutex.RLock()
	savedUserId, exists = sessionStore[cookie.Value]
	sessionMutex.RUnlock()
	assert.Equal(t, false, exists, fmt.Sprintf("Wrong exists: got %v, want %v", exists, false))
	assert.Equal(t, "", savedUserId, fmt.Sprintf("Wrong userId: got %v, want %v", savedUserId, "''"))
}

func Test_get_user_id(t *testing.T) {
	tests := []struct {
		name             string
		expectedError    error
		expectedResponse *string
	}{
		{
			name:             "Session do not exists error",
			expectedError:    errs.NewAuthenticationError("Unauthorized: No session found"),
			expectedResponse: nil,
		},
		{
			name:             "Success",
			expectedError:    nil,
			expectedResponse: &sid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, _ := setup(t, nil, nil)
			defer teardown()

			sessionToken := "mock token"
			cookie := http.Cookie{Value: sessionToken}

			if tt.expectedResponse != nil {
				sessionMutex.Lock()
				sessionStore[sessionToken] = *tt.expectedResponse
				sessionMutex.Unlock()
			}

			resp, err := mockService.GetUserId(ctx, &cookie)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedError))
		})
	}
}
