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

func setup() (func(), sqlmock.Sqlmock, UserRepository) {
	mockDB, mocksql, _ := sqlmock.New()

	client := sqlx.NewDb(mockDB, "sqlmock")
	userRepository := NewUserRepositoryDb(client)

	return func() {
		defer mockDB.Close()
	}, mocksql, userRepository
}

var exists bool = true

func Test_find_by_username(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *User
	}{
		{
			name:             "Db error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "No username found",
			dbError:          sql.ErrNoRows,
			expectedResponse: nil,
			expectedError:    errs.NewAuthenticationError("Invalid username or password"),
		},
		{
			name:             "Success",
			dbError:          nil,
			expectedResponse: &User{Id: 1},
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, userRepository := setup()
			defer teardown()

			username := "mock username"
			a := mocksql.ExpectQuery(`SELECT id, username, password FROM users`).
				WithArgs(username)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := userRepository.FindByUsername(ctx, username)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}

func Test_save(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *User
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
			expectedResponse: &User{
				Id:       1,
				Username: "mock username",
				Password: "mock password",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, userRepository := setup()
			defer teardown()

			u := User{
				Username: "mock username",
				Password: "mock password",
			}

			a := mocksql.ExpectExec(`INSERT INTO users`).
				WithArgs(u.Username, u.Password)
			if tt.dbError == nil {
				a.WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				if tt.dbError.Error() == "driver does not support LastInsertId" {
					a.WillReturnResult(sqlmock.NewErrorResult(tt.dbError))

				} else {
					a.WillReturnError(tt.dbError)
				}
			}

			res, err := userRepository.Save(ctx, u)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}

func Test_check_if_username_exists(t *testing.T) {
	tests := []struct {
		name             string
		dbError          error
		expectedError    error
		expectedResponse *bool
	}{
		{
			name:             "Db error",
			dbError:          fmt.Errorf("some error"),
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Unexpected error from database"),
		},
		{
			name:             "Success",
			dbError:          nil,
			expectedResponse: &exists,
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, mocksql, userRepository := setup()
			defer teardown()

			username := "mock username"
			a := mocksql.ExpectQuery(`SELECT EXISTS`).WithArgs(username)
			if tt.dbError == nil {
				a.WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)
			} else {
				a.WillReturnError(tt.dbError)
			}

			res, err := userRepository.CheckIfUsernameExists(ctx, username)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, res, fmt.Sprintf("Wrong response: got %v, want %v", res, tt.expectedResponse))
		})
	}
}
