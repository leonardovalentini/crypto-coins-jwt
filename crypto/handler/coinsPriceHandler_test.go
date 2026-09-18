package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/crypto/tests/mocks/service"
	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/stretchr/testify/assert"
)

func setupCoinsPrice(t *testing.T, userId *string, coin, coinRef, price, startDate, endDate, site string) (func(), *mux.Router, *service.MockCoinsPriceService) {
	ctrl := gomock.NewController(t)

	mockService := service.NewMockCoinsPriceService(ctrl)
	cph := NewCoinsPriceHandler(mockService)

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			query := r.URL.Query()
			if coin != "" {
				query.Set("coin", coin)
			}
			if coinRef != "" {
				query.Set("coin_ref", coinRef)
			}
			if price != "" {
				query.Set("price", price)
			}
			if startDate != "" {
				query.Set("start_date", startDate)
			}
			if endDate != "" {
				query.Set("end_date", endDate)
			}
			if site != "" {
				query.Set("site", site)
			}

			r.URL.RawQuery = query.Encode()

			if userId != nil {
				ctx := *contextKey.SetUserId(r, *userId)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})

	})

	jobsRouter := router.PathPrefix("/jobs").Subrouter()
	cph.RegisterPublicRoutes(jobsRouter)

	privateRouter := router.PathPrefix("/api").Subrouter()
	cph.RegisterPrivateRoutes(privateRouter)

	return func() { ctrl.Finish() }, router, mockService
}

func Test_get_all_prices_errors(t *testing.T) {
	tests := []struct {
		name                     string
		userId                   *string
		coin                     string
		coinRef                  string
		price                    string
		startDate                string
		endDate                  string
		site                     string
		gpr                      *dto.GetPricesRequest
		getPricesError           error
		getPricesResponse        []domain.Price
		getPricesTotal           *int
		getPricesResponseSummary *domain.Summary
		expectedCode             int
		expectedMessage          string
	}{
		{
			name:            "User ID not found in the request",
			userId:          nil,
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "User id not found",
		},
		{
			name:            "Price error",
			userId:          &userId,
			price:           "@",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "Invalid price value",
		},
		{
			name:            "Start date error",
			userId:          &userId,
			startDate:       "@",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "Invalid start date value, should be in ISO format",
		},
		{
			name:            "End date error",
			userId:          &userId,
			endDate:         "@",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "Invalid end date value, should be in ISO format",
		},
		{
			name:            "invalid coin error",
			userId:          &userId,
			coin:            "abcd",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin symbol must be three chars",
		},
		{
			name:            "invalid coin ref error",
			userId:          &userId,
			coinRef:         "abcd",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The coin reference symbol must be three chars",
		},
		{
			name:            "invalid price error",
			userId:          &userId,
			price:           "-10",
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The price shold be a positive number",
		},
		{
			name:            "invalid start date error",
			userId:          &userId,
			startDate:       time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The start date should be before today",
		},
		{
			name:            "invalid end date error",
			userId:          &userId,
			endDate:         time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The end date should be before today",
		},
		{
			name:            "Start date after end date error",
			userId:          &userId,
			startDate:       time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			endDate:         time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			expectedCode:    http.StatusUnprocessableEntity,
			expectedMessage: "The start date should be before end date",
		},
		{
			name:                     "Get Prices error",
			userId:                   &userId,
			gpr:                      &dto.GetPricesRequest{Page: 0, Size: 10},
			getPricesError:           errs.NewUnexpectedError("Unexpected error"),
			getPricesResponse:        nil,
			getPricesTotal:           nil,
			getPricesResponseSummary: nil,
			expectedCode:             http.StatusInternalServerError,
			expectedMessage:          "Unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown, router, mockService := setupCoinsPrice(t, tt.userId, tt.coin, tt.coinRef, tt.price, tt.startDate, tt.endDate, tt.site)
			defer teardown()

			if tt.getPricesResponse != nil || tt.getPricesResponseSummary != nil || tt.getPricesTotal != nil || tt.getPricesError != nil {
				mockService.EXPECT().GetPrices(gomock.Any(), userId, tt.gpr).Return(tt.getPricesResponse, tt.getPricesTotal, tt.getPricesResponseSummary, tt.getPricesError)
			}

			request, _ := http.NewRequest(http.MethodGet, "/api/prices", nil)

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

func Test_get_all_prices_success(t *testing.T) {
	site := "CoinGecko"
	teardown, router, mockService := setupCoinsPrice(t, &userId, "", "", "", "", "", site)
	defer teardown()

	gpr := dto.GetPricesRequest{Size: 10, Site: &site}
	getPricesResponse := []domain.Price{domain.Price{}}
	getPricesTotal := 1
	getPricesResponseSummary := domain.Summary{}
	mockService.EXPECT().GetPrices(gomock.Any(), userId, &gpr).Return(getPricesResponse, &getPricesTotal, &getPricesResponseSummary, nil)

	request, _ := http.NewRequest(http.MethodGet, "/api/prices", nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result dto.PricesResponse
	err := json.NewDecoder(resp.Body).Decode(&result)

	expectedResponse := dto.PricesResponse{
		Page:       0,
		Size:       10,
		TotalItems: 1,
		TotalPages: 1,
		Items:      []dto.PriceResponse{dto.PriceResponse{}},
		Summary:    dto.SummaryResponse{},
	}
	assert.Equal(t, nil, err, fmt.Sprintf("Wrong error: got %v, want %v", err, nil))
	assert.Equal(t, expectedResponse, result, fmt.Sprintf("Wrong response: got %v, want %v", result, expectedResponse))
}

func Test_get_broker_prices(t *testing.T) {
	teardown, router, mockService := setupCoinsPrice(t, &userId, "", "", "", "", "", "")
	defer teardown()

	mockService.EXPECT().JobPrices(gomock.Any()).Return()

	request, _ := http.NewRequest(http.MethodGet, "/jobs/broker-prices", nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	resp := recorder.Result()
	defer resp.Body.Close()

	var result struct{}
	err := json.NewDecoder(resp.Body).Decode(&result)

	expectedResponse := struct{}{}
	assert.Equal(t, nil, err, fmt.Sprintf("Wrong error: got %v, want %v", err, nil))
	assert.Equal(t, expectedResponse, result, fmt.Sprintf("Wrong response: got %v, want %v", result, expectedResponse))
}
