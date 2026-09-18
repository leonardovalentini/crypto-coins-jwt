package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/stretchr/testify/assert"
)

var total int = 1

func setupPriceTests() (func(), sqlmock.Sqlmock, PriceRepository) {
	mockDB, mocksql, _ := sqlmock.New()

	client := sqlx.NewDb(mockDB, "sqlmock")
	priceRepository := NewPriceRepositoryDb(client)

	return func() {
		defer mockDB.Close()
	}, mocksql, priceRepository
}

func Test_find_all_by(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse []Price
	}{
		{
			name:             "Db error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "No Prices matching filter",
			dbError:          nil,
			expectedResponse: []Price{},
			expectedError:    nil,
		},
		{
			name:             "Success",
			dbError:          nil,
			expectedResponse: []Price{Price{Id: 1}},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, priceRepository := setupPriceTests()
			defer teardown()

			userId := "1"
			coin := "btc"
			coinRef := "usd"
			price := 1.0
			startDate := time.Now()
			endDate := time.Now()
			site := "site"

			gpr := dto.GetPricesRequest{
				Page:      0,
				Size:      10,
				Coin:      &coin,
				CoinRef:   &coinRef,
				Price:     &price,
				StartDate: &startDate,
				EndDate:   &endDate,
				Site:      &site,
			}

			a := mocksql.ExpectQuery(`SELECT p.coin, p.coin_ref, p.date, p.price, p.site FROM prices AS p INNER JOIN user_coins AS uc`).
				WithArgs(userId, gpr.Coin, gpr.CoinRef, gpr.Price, gpr.StartDate, gpr.EndDate, gpr.Site, gpr.Size, gpr.Page*gpr.Size)
			if tt.dbError == nil {
				if len(tt.expectedResponse) == 0 {
					a.WillReturnRows(
						sqlmock.NewRows([]string{"id"}),
					)
				} else {
					a.WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(1),
					)
				}
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := priceRepository.FindAllBy(ctx, userId, &gpr)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}

func Test_get_summary(t *testing.T) {
	tests := []struct {
		name            string
		dbError         error
		expectedError   error
		expectedTotal   *int
		expectedSummary *Summary
	}{
		{
			name:            "Db error",
			dbError:         fmt.Errorf("some error"),
			expectedTotal:   nil,
			expectedSummary: nil,
			expectedError:   errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:            "Success",
			dbError:         nil,
			expectedTotal:   &total,
			expectedSummary: &Summary{Coin: map[string]int{"btc": 1}, CoinRef: map[string]int{"usd": 1}, Site: map[string]int{"CoinGecko": 1}},
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, priceRepository := setupPriceTests()
			defer teardown()

			userId := "1"
			coin := "btc"
			coinRef := "usd"
			price := 1.0
			startDate := time.Now()
			endDate := time.Now()
			site := "site"

			gpr := dto.GetPricesRequest{
				Page:      0,
				Size:      10,
				Coin:      &coin,
				CoinRef:   &coinRef,
				Price:     &price,
				StartDate: &startDate,
				EndDate:   &endDate,
				Site:      &site,
			}

			a := mocksql.ExpectQuery(`SELECT "coin" AS column_name, p.coin AS value`).
				WithArgs(userId, gpr.Coin, gpr.CoinRef, gpr.Price, gpr.StartDate, gpr.EndDate, gpr.Site, userId, gpr.Coin, gpr.CoinRef, gpr.Price, gpr.StartDate, gpr.EndDate, gpr.Site, userId, gpr.Coin, gpr.CoinRef, gpr.Price, gpr.StartDate, gpr.EndDate, gpr.Site)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"column_name", "value", "count"}).AddRow("coin", "btc", 1).AddRow("coin_ref", "usd", 1).AddRow("site", "CoinGecko", 1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			total, summary, err := priceRepository.GetSummary(ctx, userId, &gpr)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedTotal, total, fmt.Sprintf("Wrong response: got %v, want %v", total, tt.expectedTotal))
			assert.Equal(t, tt.expectedSummary, summary, fmt.Sprintf("Wrong response: got %v, want %v", summary, tt.expectedSummary))
		})
	}
}

func Test_price_save(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *Price
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
			name:    "Success",
			dbError: nil,
			expectedResponse: &Price{
				Id:      1,
				Coin:    "btc",
				CoinRef: "usd",
				Date:    time.Now(),
				Price:   1,
				Site:    "CoinMarketCap"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, priceRepository := setupPriceTests()
			defer teardown()

			p := Price{
				Coin:    "btc",
				CoinRef: "usd",
				Date:    time.Now(),
				Price:   1,
				Site:    "CoinMarketCap",
			}
			if tt.expectedResponse != nil {
				p.Date = tt.expectedResponse.Date
			}

			a := mocksql.ExpectExec(`INSERT INTO prices`).
				WithArgs(p.Coin, p.CoinRef, p.Date, p.Price, p.Site)
			if tt.dbError == nil {
				a.WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				if tt.dbError.Error() == "driver does not support LastInsertId" {
					a.WillReturnResult(sqlmock.NewErrorResult(tt.dbError))

				} else {
					a.WillReturnError(tt.dbError)
				}
			}

			res, err := priceRepository.Save(ctx, p)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}
