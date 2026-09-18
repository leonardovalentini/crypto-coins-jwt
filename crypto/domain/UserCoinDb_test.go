package domain

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/stretchr/testify/assert"
)

func setup() (func(), sqlmock.Sqlmock, UserCoinRepository) {
	mockDB, mocksql, _ := sqlmock.New()

	client := sqlx.NewDb(mockDB, "sqlmock")
	UserCoinRepo := NewUserCoinRepositoryDb(client)

	return func() {
		defer mockDB.Close()
	}, mocksql, UserCoinRepo
}

func Test_find_all_by_user_id(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse []UserCoin
	}{
		{
			name:             "Db error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Error while querying UserCoins table some error"),
		},
		{
			name:             "Success",
			expectedResponse: []UserCoin{UserCoin{Id: 1, UserId: "", Coin: "", CoinRef: ""}},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()

			id := "1"

			a := mocksql.ExpectQuery(`SELECT id, coin, coin_ref FROM user_coins`).
				WithArgs(id)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := UserCoinRepo.FindAllByUserId(ctx, id)

			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_find_by(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse []UserCoin
	}{
		{
			name:             "Db error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Error while querying UserCoins table some error"),
		},
		{
			name:             "Success",
			expectedResponse: []UserCoin{UserCoin{Id: 1, UserId: "", Coin: "", CoinRef: ""}},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()

			id := "1"
			coin := "btc"
			coinRef := "usd"

			a := mocksql.ExpectQuery(`SELECT id, coin, coin_ref FROM user_coins`).
				WithArgs(id, coin, coinRef)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := UserCoinRepo.FindBy(ctx, id, coin, coinRef)

			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_find_by_user_coin_id(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *UserCoin
	}{
		{
			name:             "Db unexpected error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Error while querying UserCoins table some error"),
		},
		{
			name:             "Db no rows error",
			dbError:          sql.ErrNoRows,
			expectedResponse: nil,
			expectedError:    errs.NewNotFoundError("The coin was not found"),
		},
		{
			name:             "Success",
			expectedResponse: &UserCoin{Id: 1, UserId: "", Coin: "", CoinRef: ""},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()

			id := "1"
			userCoinId := "1"

			a := mocksql.ExpectQuery(`SELECT id, user_id, coin, coin_ref FROM user_coins`).
				WithArgs(id, userCoinId)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := UserCoinRepo.FindByUserCoinId(ctx, id, userCoinId)

			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_find_differents_coins(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse []UserCoin
	}{
		{
			name:             "Db unexpected error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Error while querying UserCoins table some error"),
		},
		{
			name:             "Success",
			expectedResponse: []UserCoin{UserCoin{Id: 1, UserId: "", Coin: "", CoinRef: ""}},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()

			a := mocksql.ExpectQuery(`SELECT DISTINCT coin, coin_ref FROM user_coins`).
				WithArgs()
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := UserCoinRepo.FindDifferentsCoins(ctx)

			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_save(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *UserCoin
	}{
		{
			name:             "Db unexpected error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "Db driver does not support LastInsertId",
			dbError:          fmt.Errorf("driver does not support LastInsertId"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "Success",
			expectedResponse: &UserCoin{Id: 1, UserId: "1", Coin: "btc", CoinRef: "usd"},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()
			u := UserCoin{
				UserId:  "1",
				Coin:    "btc",
				CoinRef: "usd",
			}

			a := mocksql.ExpectExec(`INSERT INTO user_coins`).
				WithArgs(u.UserId, u.Coin, u.CoinRef)
			if tt.dbError == nil {
				a.WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				if tt.dbError.Error() == "driver does not support LastInsertId" {
					a.WillReturnResult(sqlmock.NewErrorResult(tt.dbError))

				} else {
					a.WillReturnError(tt.dbError)
				}
			}

			res, err := UserCoinRepo.Save(ctx, u)

			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}

func Test_update(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *UserCoin
	}{
		{
			name:             "Db unexpected error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "Success",
			expectedResponse: &UserCoin{Id: 1, UserId: "1", Coin: "btc", CoinRef: "usd"},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()
			u := UserCoin{
				Id:      1,
				UserId:  "1",
				Coin:    "btc",
				CoinRef: "usd",
			}

			a := mocksql.ExpectExec(`UPDATE user_coins`).
				WithArgs(u.Coin, u.CoinRef, u.Id, u.UserId)
			if tt.dbError == nil {
				a.WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := UserCoinRepo.Update(ctx, u)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}

func Test_delete(t *testing.T) {
	tests := []struct {
		name          string
		dbError       error
		expectedError error
	}{
		{
			name:          "Db unexpected error",
			dbError:       fmt.Errorf("some error"),
			expectedError: errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:          "Success",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, UserCoinRepo := setup()
			defer teardown()
			userCoinId := "1"

			a := mocksql.ExpectExec(`DELETE FROM user_coins`).
				WithArgs(userCoinId)
			if tt.dbError == nil {
				a.WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				a.WillReturnError(tt.dbError)
			}

			err := UserCoinRepo.Delete(ctx, userCoinId)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
		})
	}
}
