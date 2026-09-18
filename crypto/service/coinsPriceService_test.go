package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	domainMock "github.com/leonardovalentini/crypto-coins/crypto/tests/mocks/domain"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	loggerMock "github.com/leonardovalentini/crypto-coins/lib/tests/mocks/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type CustomError struct {
	Code    int    `json:",omitempty"`
	Message string `json:"message"`
}

func (e CustomError) Error() string {
	panic("unhandled error")
}

var getCurrentTime = func() time.Time { return time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC) }
var total int = 1
var price float64 = 10.0

func setupCoinsPrice(t *testing.T) (func(), CoinsPriceService, *domainMock.MockUserCoinRepository, *domainMock.MockPriceRepository, *domainMock.MockBrokerSite, *loggerMock.MockLoggerI, func() time.Time) {
	ctrl := gomock.NewController(t)
	mockUserCoinRepo := domainMock.NewMockUserCoinRepository(ctrl)
	mockPriceRepo := domainMock.NewMockPriceRepository(ctrl)
	mockBrockerRepo := domainMock.NewMockBrokerSite(ctrl)

	mockLogger := loggerMock.NewMockLoggerI(ctrl)

	mockService := NewCoinsPriceService(mockUserCoinRepo, mockPriceRepo, []domain.BrokerSite{mockBrockerRepo}, mockLogger, getCurrentTime)
	return func() {
		mockService = nil
		defer ctrl.Finish()
	}, mockService, mockUserCoinRepo, mockPriceRepo, mockBrockerRepo, mockLogger, getCurrentTime
}

func Test_get_prices(t *testing.T) {
	tests := []struct {
		name               string
		findAllByError     error
		findAllByResponse  []domain.Price
		getSummaryError    error
		getSummaryResponse *domain.Summary
		getTotalResponse   *int
		expectedError      error
		expectedResponse   []domain.Price
		expectedTotal      *int
		expectedSummary    *domain.Summary
	}{
		{
			name:              "Find all by unexpected error",
			findAllByResponse: nil,
			findAllByError:    errs.NewUnexpectedError("unexpected error"),
			expectedError:     errs.NewUnexpectedError("unexpected error"),
			expectedResponse:  nil,
			expectedTotal:     nil,
			expectedSummary:   nil,
		},
		{
			name: "Get summary unexpected error",
			findAllByResponse: []domain.Price{domain.Price{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
				Date:    getCurrentTime(),
				Price:   10.0,
				Site:    "CoinGecko",
			}},
			findAllByError:     nil,
			getSummaryError:    errs.NewUnexpectedError("unexpected error"),
			getSummaryResponse: nil,
			getTotalResponse:   nil,
			expectedError:      errs.NewUnexpectedError("unexpected error"),
			expectedResponse:   nil,
			expectedTotal:      nil,
			expectedSummary:    nil,
		},
		{
			name: "Success",
			findAllByResponse: []domain.Price{domain.Price{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
				Date:    getCurrentTime(),
				Price:   10.0,
				Site:    "CoinGecko",
			}},
			findAllByError:  nil,
			getSummaryError: nil,
			getSummaryResponse: &domain.Summary{
				Coin:    map[string]int{"btc": 1},
				CoinRef: map[string]int{"usd": 1},
				Site:    map[string]int{"CoinGecko": 1},
			},
			getTotalResponse: &total,
			expectedError:    nil,
			expectedResponse: []domain.Price{domain.Price{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
				Date:    getCurrentTime(),
				Price:   10.0,
				Site:    "CoinGecko",
			}},
			expectedTotal: &total,
			expectedSummary: &domain.Summary{
				Coin:    map[string]int{"btc": 1},
				CoinRef: map[string]int{"usd": 1},
				Site:    map[string]int{"CoinGecko": 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, _, mockPriceRepo, _, _, _ := setupCoinsPrice(t)
			defer teardown()

			userId := "1"
			gpr := dto.GetPricesRequest{}

			if tt.findAllByResponse != nil || tt.findAllByError != nil {
				mockPriceRepo.EXPECT().FindAllBy(ctx, userId, &gpr).Return(tt.findAllByResponse, tt.findAllByError)
			}

			if tt.getSummaryResponse != nil || tt.getSummaryError != nil {
				mockPriceRepo.EXPECT().GetSummary(ctx, userId, &gpr).Return(tt.getTotalResponse, tt.getSummaryResponse, tt.getSummaryError)
			}

			resp, total, summary, err := mockService.GetPrices(ctx, userId, &gpr)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedError))
			assert.Equal(t, tt.expectedTotal, total, fmt.Sprintf("Wrong total: got %v, want %v", total, tt.expectedError))
			assert.Equal(t, tt.expectedSummary, summary, fmt.Sprintf("Wrong summary: got %v, want %v", summary, tt.expectedError))
		})
	}
}

func Test_job_prices(t *testing.T) {
	tests := []struct {
		name                        string
		findDifferentsCoinsError    error
		findDifferentsCoinsResponse []domain.UserCoin
		getPriceError               error
		getPriceResponse            *float64
		saveError                   error
		saveResponse                *domain.Price
		message                     string
		messageSuccess              string
	}{
		{
			name:                        "Job prices unhandled error",
			findDifferentsCoinsResponse: nil,
			findDifferentsCoinsError:    CustomError{},
			message:                     "Encountered unexpected error: unhandled error",
		},
		{
			name:                        "DB: unexpected error in find differents coins",
			findDifferentsCoinsResponse: nil,
			findDifferentsCoinsError:    errs.NewUnexpectedError("Unexpected error"),
			message:                     "Error in JobPrices: Unexpected error",
		},
		{
			name: "Get prices by site and coin unhandled error",
			findDifferentsCoinsResponse: []domain.UserCoin{domain.UserCoin{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
			}},
			findDifferentsCoinsError: nil,
			getPriceError:            CustomError{},
			getPriceResponse:         nil,
			message:                  "Encountered unexpected error: unhandled error",
		},
		{
			name: "DB: unexpected error in get price",
			findDifferentsCoinsResponse: []domain.UserCoin{domain.UserCoin{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
			}},
			findDifferentsCoinsError: nil,
			getPriceError:            errs.NewUnexpectedError("Unexpected error"),
			getPriceResponse:         nil,
			message:                  "JobPrices: Unexpected error getting the price Unexpected error",
		},
		{
			name: "DB: unexpected error in save",
			findDifferentsCoinsResponse: []domain.UserCoin{domain.UserCoin{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
			}},
			findDifferentsCoinsError: nil,
			getPriceError:            nil,
			getPriceResponse:         &price,
			saveError:                errs.NewUnexpectedError("Unexpected error"),
			saveResponse:             nil,
			message:                  "JobPrices: Unexpected error saving the price: Unexpected error",
		},
		{
			name: "Success",
			findDifferentsCoinsResponse: []domain.UserCoin{domain.UserCoin{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
			}},
			findDifferentsCoinsError: nil,
			getPriceError:            nil,
			getPriceResponse:         &price,
			saveError:                nil,
			saveResponse:             &domain.Price{},
			messageSuccess:           "JobPrices: Price saved successfully for btc/usd from Mock Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockUserCoinRepo, mockPriceRepo, mockBrockerRepo, mockLogger, getCurrentTime := setupCoinsPrice(t)
			defer teardown()

			if tt.findDifferentsCoinsResponse != nil || tt.findDifferentsCoinsError != nil {
				mockUserCoinRepo.EXPECT().FindDifferentsCoins(gomock.Any()).
					Return(tt.findDifferentsCoinsResponse, tt.findDifferentsCoinsError)
			}

			if tt.getPriceResponse != nil || tt.getPriceError != nil {
				mockBrockerRepo.EXPECT().GetPrice(gomock.Any(), "btc", "usd").
					Return(tt.getPriceResponse, tt.getPriceError)
			}

			if tt.saveResponse != nil || tt.saveError != nil {
				mockBrockerRepo.EXPECT().GetName().
					Return("Mock Name").AnyTimes()

				p := domain.Price{
					Coin:    "btc",
					CoinRef: "usd",
					Date:    getCurrentTime(),
					Price:   price,
					Site:    "Mock Name",
				}
				mockPriceRepo.EXPECT().Save(ctx, p).
					Return(tt.saveResponse, tt.saveError)

			}

			mockLogger.EXPECT().SetCtx(ctx).Return().AnyTimes()
			if tt.message != "" {
				mockLogger.EXPECT().Error(tt.message).Return()
			}
			if tt.messageSuccess != "" {
				mockLogger.EXPECT().Info(tt.messageSuccess).Return()
			}
			mockService.JobPrices(ctx)
		})
	}
}
