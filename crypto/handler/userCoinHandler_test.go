package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/crypto/tests/mocks/service"
	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
)

var userId string = "1"

func setup(t *testing.T, userId *string) (func(), *mux.Router, *service.MockUserCoinService) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockUserCoinService(ctrl)
	ch := NewUserCoinHandler(mockService)

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userId != nil {
				ctx := *contextKey.SetUserId(r, *userId)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})

	})

	privateRouter := router.PathPrefix("/api").Subrouter()
	ch.RegisterPrivateRoutes(privateRouter)

	return func() { ctrl.Finish() }, router, mockService
}

func Test_get_all_user_coins_Errors(t *testing.T) {
	tests := []struct {
		name            string
		userId          *string
		expectedCode    int
		expectedMessage string
		serviceError    error
	}{
		{
			name:            "Get user-coins: User ID not found in the request",
			userId:          nil,
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "User id not found",
		},
		{
			name:            "Get user-coins: Unexpected error",
			userId:          &userId,
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: "Error while querying UserCoins table",
			serviceError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t, tt.userId)
			defer teardown()

			if tt.serviceError != nil {
				mockService.EXPECT().GetAllUserCoinsByUserId(gomock.Any(), *tt.userId).Return(nil, tt.serviceError)
			}

			request, _ := http.NewRequest(http.MethodGet, "/api/user-coins", nil)

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

func Test_get_all_user_coins_success(t *testing.T) {
	teardown, router, mockService := setup(t, &userId)
	defer teardown()

	mockServiceResponse := []domain.UserCoin{
		domain.UserCoin{
			Id:      1,
			UserId:  userId,
			Coin:    "btc",
			CoinRef: "usd",
		},
	}
	expectedServiceResponse := []dto.UserCoinResponse{
		dto.UserCoinResponse{
			Id:      1,
			Coin:    "btc",
			CoinRef: "usd",
		},
	}

	mockService.EXPECT().GetAllUserCoinsByUserId(gomock.Any(), userId).Return(mockServiceResponse, nil)

	request, _ := http.NewRequest(http.MethodGet, "/api/user-coins", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result []dto.UserCoinResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}

	assert.Equal(t, http.StatusOK, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.ElementsMatch(t, expectedServiceResponse, result, fmt.Sprintf("Wrong response body: got %v, want %v", result, expectedServiceResponse))
}

func Test_new_user_coin_Errors(t *testing.T) {
	tests := []struct {
		name            string
		userId          *string
		body            any
		expectedCode    int
		expectedMessage string
		serviceError    error
	}{
		{
			name:            "Auth: User ID not found in the request",
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "User id not found",
		},
		{
			name:            "Body: invalid json",
			userId:          &userId,
			body:            "<invalid json>",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Invalid request body",
		},
		{
			name:            "Body: invalid values",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btcc", CoinRef: "usd"},
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin symbol must be three chars",
			serviceError:    errs.NewValidationError("The coin symbol must be three chars"),
		},
		{
			name:            "Body: invalid values",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "us"},
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin reference must be three chars",
			serviceError:    errs.NewValidationError("The coin reference must be three chars"),
		},
		{
			name:            "DB: Unexpected Error",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			expectedCode:    http.StatusConflict,
			expectedMessage: "The user has already added that coin",
			serviceError:    errs.NewConflictError("The user has already added that coin"),
		},
		{
			name:            "DB: Unexpected Error",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: "Error while querying UserCoins table",
			serviceError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t, tt.userId)
			defer teardown()

			if tt.serviceError != nil {
				mockService.EXPECT().NewUserCoin(gomock.Any(), *tt.userId, tt.body).Return(nil, tt.serviceError)
			}

			out, err := json.Marshal(tt.body)
			if err != nil {
				assert.Error(t, err, "Error in json Marshal")
			}
			request, _ := http.NewRequest(http.MethodPost, "/api/user-coins", bytes.NewBuffer(out))

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

func Test_new_user_coin_success(t *testing.T) {
	teardown, router, mockService := setup(t, &userId)
	defer teardown()

	body := dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"}
	response := dto.NewUserCoinResponse{Id: 1}
	mockService.EXPECT().NewUserCoin(gomock.Any(), userId, body).Return(&response, nil)

	out, err := json.Marshal(body)
	if err != nil {
		assert.Error(t, err, "Error in json Marshal")
	}
	request, _ := http.NewRequest(http.MethodPost, "/api/user-coins", bytes.NewBuffer(out))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.NewUserCoinResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}

	assert.Equal(t, http.StatusCreated, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.Equal(t, response, result, fmt.Sprintf("Wrong response body: got %v, want %v", result, response))
}

func Test_update_user_coin_Errors(t *testing.T) {
	tests := []struct {
		name            string
		userId          *string
		body            any
		expectedCode    int
		expectedMessage string
		serviceError    error
	}{
		{
			name:            "Auth: User ID not found in the request",
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "User id not found",
		},
		{
			name:            "Body: invalid json",
			userId:          &userId,
			body:            "<invalid json>",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Invalid request body",
		},
		{
			name:            "Body: invalid values",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btcc", CoinRef: "usd"},
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin symbol must be three chars",
			serviceError:    errs.NewValidationError("The coin symbol must be three chars"),
		},
		{
			name:            "Body: invalid values",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "us"},
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin reference must be three chars",
			serviceError:    errs.NewValidationError("The coin reference must be three chars"),
		},
		{
			name:            "DB: Coin not found",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			expectedCode:    http.StatusNotFound,
			expectedMessage: "The coin was not found",
			serviceError:    errs.NewNotFoundError("The coin was not found"),
		},
		{
			name:            "DB: Coin duplicated",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			expectedCode:    http.StatusConflict,
			expectedMessage: "The user has already added that coin",
			serviceError:    errs.NewConflictError("The user has already added that coin"),
		},
		{
			name:            "DB: Unexpected Error",
			userId:          &userId,
			body:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: "Error while querying UserCoins table",
			serviceError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t, tt.userId)
			defer teardown()

			if tt.serviceError != nil {
				mockService.EXPECT().UpdateUserCoin(gomock.Any(), *tt.userId, "1", tt.body).Return(nil, tt.serviceError)
			}

			out, err := json.Marshal(tt.body)
			if err != nil {
				assert.Error(t, err, "Error in json Marshal")
			}
			request, _ := http.NewRequest(http.MethodPut, "/api/user-coins/1", bytes.NewBuffer(out))

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

func Test_update_user_coin_success(t *testing.T) {
	teardown, router, mockService := setup(t, &userId)
	defer teardown()

	body := dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"}
	response := domain.UserCoin{Id: 1, UserId: userId, Coin: "btc", CoinRef: "usd"}
	mockService.EXPECT().UpdateUserCoin(gomock.Any(), userId, "1", body).Return(&response, nil)

	out, err := json.Marshal(body)
	if err != nil {
		assert.Error(t, err, "Error in json Marshal")
	}
	request, _ := http.NewRequest(http.MethodPut, "/api/user-coins/1", bytes.NewBuffer(out))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.UserCoinResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}
	expectedResult := dto.UserCoinResponse{Id: 1, Coin: "btc", CoinRef: "usd"}

	assert.Equal(t, http.StatusOK, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.Equal(t, expectedResult, result, fmt.Sprintf("Wrong response body: got %v, want %v", result, expectedResult))
}

func Test_delete_user_coin_Errors(t *testing.T) {
	tests := []struct {
		name            string
		userId          *string
		expectedCode    int
		expectedMessage string
		serviceError    error
	}{
		{
			name:            "Auth: User ID not found in the request",
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "User id not found",
		},
		{
			name:            "DB: Coin not found",
			userId:          &userId,
			expectedCode:    http.StatusNotFound,
			expectedMessage: "The coin was not found",
			serviceError:    errs.NewNotFoundError("The coin was not found"),
		},
		{
			name:            "DB: Unexpected Error",
			userId:          &userId,
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: "Error while querying UserCoins table",
			serviceError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setup(t, tt.userId)
			defer teardown()

			if tt.serviceError != nil {
				mockService.EXPECT().DeleteUserCoin(gomock.Any(), *tt.userId, "1").Return(tt.serviceError)
			}

			request, _ := http.NewRequest(http.MethodDelete, "/api/user-coins/1", nil)

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

func Test_delete_user_coin_success(t *testing.T) {
	teardown, router, mockService := setup(t, &userId)
	defer teardown()

	mockService.EXPECT().DeleteUserCoin(gomock.Any(), userId, "1").Return(nil)

	request, _ := http.NewRequest(http.MethodDelete, "/api/user-coins/1", nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.EmptyResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		assert.Error(t, err, "Error reading body")
	}
	expectedResult := dto.EmptyResponse{}

	assert.Equal(t, http.StatusOK, recorder.Code, fmt.Sprintf("Wrong response code: got %d, want %d", recorder.Code, http.StatusOK))
	assert.Equal(t, expectedResult, result, fmt.Sprintf("Wrong response body: got %v, want %v", result, expectedResult))
}
