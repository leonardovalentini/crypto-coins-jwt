package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	domainMock "github.com/leonardovalentini/crypto-coins/crypto/tests/mocks/domain"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const userId string = "1"
const userCoinId string = "1"

func setup(t *testing.T) (func(), UserCoinService, *domainMock.MockUserCoinRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := domainMock.NewMockUserCoinRepository(ctrl)
	mockService := NewUserCoinService(mockRepo)
	return func() {
		mockService = nil
		defer ctrl.Finish()
	}, mockService, mockRepo
}

func Test_get_all_user_coins_by_user_id(t *testing.T) {
	tests := []struct {
		name         string
		repoError    error
		repoResponse []domain.UserCoin
	}{
		{
			name:         "DB: unexpected error",
			repoResponse: nil,
			repoError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
		},
		{
			name: "Success",
			repoResponse: []domain.UserCoin{
				domain.UserCoin{
					Id:      1,
					UserId:  userId,
					Coin:    "btc",
					CoinRef: "usd",
				},
			},
			repoError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t)
			defer teardown()

			mockRepo.EXPECT().FindAllByUserId(ctx, userId).Return(tt.repoResponse, tt.repoError)
			resp, err := mockService.GetAllUserCoinsByUserId(ctx, userId)

			assert.Equal(t, tt.repoResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.repoResponse))
			assert.Equal(t, tt.repoError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.repoError))
		})
	}
}

func Test_new_user_coin(t *testing.T) {
	tests := []struct {
		name             string
		req              dto.UserCoinRequest
		findByError      error
		findByResponse   []domain.UserCoin
		saveError        error
		saveResponse     *domain.UserCoin
		expectedError    error
		expectedResponse *dto.NewUserCoinResponse
	}{
		{
			name:             "Validation error coin",
			req:              dto.UserCoinRequest{Coin: "btcc", CoinRef: "usd"},
			expectedError:    errs.NewValidationError("The coin symbol must be three chars"),
			expectedResponse: nil,
		},
		{
			name:             "Validation error coin ref",
			req:              dto.UserCoinRequest{Coin: "btc", CoinRef: "us"},
			expectedError:    errs.NewValidationError("The coin reference must be three chars"),
			expectedResponse: nil,
		},
		{
			name:             "Repo find by error",
			req:              dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByError:      errs.NewUnexpectedError("Error while querying UserCoins table"),
			expectedError:    errs.NewUnexpectedError("Error while querying UserCoins table"),
			expectedResponse: nil,
		},
		{
			name:             "Duplicated coin",
			req:              dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByError:      nil,
			findByResponse:   []domain.UserCoin{domain.UserCoin{}},
			expectedError:    errs.NewConflictError("The user has already added that coin"),
			expectedResponse: nil,
		},
		{
			name:             "Repo save error",
			req:              dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByError:      nil,
			findByResponse:   []domain.UserCoin{},
			saveError:        errs.NewUnexpectedError("Unexpected error from database"),
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
			expectedResponse: nil,
		},
		{
			name:           "Success",
			req:            dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByError:    nil,
			findByResponse: []domain.UserCoin{},
			saveError:      nil,
			saveResponse: &domain.UserCoin{
				Id:      2,
				UserId:  userId,
				Coin:    "btc",
				CoinRef: "usd",
			},
			expectedError:    nil,
			expectedResponse: &dto.NewUserCoinResponse{Id: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t)
			defer teardown()

			userCoin := domain.UserCoin{
				UserId:  userId,
				Coin:    tt.req.Coin,
				CoinRef: tt.req.CoinRef,
			}
			if tt.findByResponse != nil || tt.findByError != nil {
				mockRepo.EXPECT().FindBy(ctx, userId, tt.req.Coin, tt.req.CoinRef).Return(tt.findByResponse, tt.findByError)
			}
			if tt.saveResponse != nil || tt.saveError != nil {
				mockRepo.EXPECT().Save(ctx, userCoin).Return(tt.saveResponse, tt.saveError)
			}

			resp, err := mockService.NewUserCoin(ctx, userId, tt.req)

			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_update_user_coin(t *testing.T) {
	updatedResponse := &domain.UserCoin{
		Id:      1,
		UserId:  userId,
		Coin:    "btc",
		CoinRef: "usd",
	}

	tests := []struct {
		name                     string
		req                      dto.UserCoinRequest
		findByUserCoinIdError    error
		findByUserCoinIdResponse *domain.UserCoin
		findByError              error
		findByResponse           []domain.UserCoin
		updateError              error
		updateResponse           *domain.UserCoin
		expectedError            error
		expectedResponse         *domain.UserCoin
	}{
		{
			name:             "Validation error coin",
			req:              dto.UserCoinRequest{Coin: "btcc", CoinRef: "usd"},
			expectedError:    errs.NewValidationError("The coin symbol must be three chars"),
			expectedResponse: nil,
		},
		{
			name:             "Validation error coin ref",
			req:              dto.UserCoinRequest{Coin: "btc", CoinRef: "us"},
			expectedError:    errs.NewValidationError("The coin reference must be three chars"),
			expectedResponse: nil,
		},
		{
			name:                  "Repo find by coin id error",
			req:                   dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError: errs.NewNotFoundError("The coin was not found"),
			expectedError:         errs.NewNotFoundError("The coin was not found"),
			expectedResponse:      nil,
		},
		{
			name:                     "Success: Nothing to update",
			req:                      dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError:    nil,
			findByUserCoinIdResponse: &domain.UserCoin{Coin: "btc", CoinRef: "usd"},
			expectedError:            nil,
			expectedResponse: &domain.UserCoin{
				Id:      0,
				UserId:  "",
				Coin:    "btc",
				CoinRef: "usd",
			},
		},
		{
			name:                     "Repo find by error",
			req:                      dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError:    nil,
			findByUserCoinIdResponse: &domain.UserCoin{Coin: "btc", CoinRef: "eth"},
			findByError:              errs.NewUnexpectedError("Error while querying UserCoins table"),
			expectedError:            errs.NewUnexpectedError("Error while querying UserCoins table"),
			expectedResponse:         nil,
		},
		{
			name:                     "Duplicated coin",
			req:                      dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError:    nil,
			findByUserCoinIdResponse: &domain.UserCoin{Coin: "btc", CoinRef: "eth"},
			findByError:              nil,
			findByResponse:           []domain.UserCoin{domain.UserCoin{}},
			expectedError:            errs.NewConflictError("The user has already added that coin"),
			expectedResponse:         nil,
		},
		{
			name:                     "Repo update error",
			req:                      dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError:    nil,
			findByUserCoinIdResponse: &domain.UserCoin{Id: 1, UserId: userId, Coin: "btc", CoinRef: "eth"},
			findByError:              nil,
			findByResponse:           []domain.UserCoin{},
			updateError:              errs.NewUnexpectedError("Unexpected error from database"),
			expectedError:            errs.NewUnexpectedError("Unexpected error from database"),
			expectedResponse:         nil,
		},
		{
			name:                     "Success",
			req:                      dto.UserCoinRequest{Coin: "btc", CoinRef: "usd"},
			findByUserCoinIdError:    nil,
			findByUserCoinIdResponse: &domain.UserCoin{Id: 1, UserId: userId, Coin: "btc", CoinRef: "eth"},
			findByError:              nil,
			findByResponse:           []domain.UserCoin{},
			updateError:              nil,
			updateResponse:           updatedResponse,
			expectedError:            nil,
			expectedResponse:         updatedResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t)
			defer teardown()

			userCoin := domain.UserCoin{
				Id:      1,
				UserId:  userId,
				Coin:    tt.req.Coin,
				CoinRef: tt.req.CoinRef,
			}
			if tt.findByUserCoinIdResponse != nil || tt.findByUserCoinIdError != nil {
				mockRepo.EXPECT().FindByUserCoinId(ctx, userId, userCoinId).Return(tt.findByUserCoinIdResponse, tt.findByUserCoinIdError)
			}
			if tt.findByResponse != nil || tt.findByError != nil {
				mockRepo.EXPECT().FindBy(ctx, userId, tt.req.Coin, tt.req.CoinRef).Return(tt.findByResponse, tt.findByError)
			}
			if tt.updateResponse != nil || tt.updateError != nil {
				mockRepo.EXPECT().Update(ctx, userCoin).Return(tt.updateResponse, tt.updateError)
			}

			resp, err := mockService.UpdateUserCoin(ctx, userId, userCoinId, tt.req)

			assert.Equal(t, tt.expectedResponse, resp, fmt.Sprintf("Wrong response: got %v, want %v", resp, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_delete_user_coin(t *testing.T) {
	tests := []struct {
		name                     string
		findByUserCoinIdError    error
		findByUserCoinIdResponse *domain.UserCoin
		deleteError              error
		expectedError            error
	}{

		{
			name:                  "User coin id not found",
			findByUserCoinIdError: errs.NewNotFoundError("The coin was not found"),
			expectedError:         errs.NewNotFoundError("The coin was not found"),
		},
		{
			name:                  "Delete error",
			findByUserCoinIdError: nil,
			findByUserCoinIdResponse: &domain.UserCoin{
				Id:      1,
				UserId:  userId,
				Coin:    "btc",
				CoinRef: "usd",
			},
			deleteError:   errs.NewUnexpectedError("Unexpected error from database"),
			expectedError: errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:                  "Success",
			findByUserCoinIdError: nil,
			findByUserCoinIdResponse: &domain.UserCoin{
				Id:      1,
				UserId:  userId,
				Coin:    "btc",
				CoinRef: "usd",
			},
			deleteError:   nil,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mockService, mockRepo := setup(t)
			defer teardown()

			if tt.findByUserCoinIdResponse != nil || tt.findByUserCoinIdError != nil {
				mockRepo.EXPECT().FindByUserCoinId(ctx, userId, userCoinId).Return(tt.findByUserCoinIdResponse, tt.findByUserCoinIdError)
			}
			if tt.deleteError != nil || tt.name == "Success" {
				mockRepo.EXPECT().Delete(ctx, userCoinId).Return(tt.deleteError)
			}

			err := mockService.DeleteUserCoin(ctx, userId, userCoinId)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}
