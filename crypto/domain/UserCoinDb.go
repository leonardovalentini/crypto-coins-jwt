package domain

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type UserCoinRepositoryDb struct {
	client *sqlx.DB
}

func NewUserCoinRepositoryDb(client *sqlx.DB) UserCoinRepository {
	return &UserCoinRepositoryDb{client: client}
}

func (c *UserCoinRepositoryDb) FindAllByUserId(ctx context.Context, id string) ([]UserCoin, error) {
	var err error
	UserCoins := make([]UserCoin, 0)

	query := "SELECT id, coin, coin_ref FROM user_coins WHERE user_id = ?"

	err = c.client.SelectContext(ctx, &UserCoins, query, id)

	if err != nil {
		return nil, errs.NewUnexpectedError("Error while querying UserCoins table " + err.Error())
	}

	return UserCoins, nil
}

func (c *UserCoinRepositoryDb) FindBy(ctx context.Context, userId, coin, coinRef string) ([]UserCoin, error) {
	var err error
	userCoins := make([]UserCoin, 0)

	query := "SELECT id, coin, coin_ref FROM user_coins WHERE user_id = ? AND coin = ? and coin_ref = ?"

	err = c.client.SelectContext(ctx, &userCoins, query, userId, coin, coinRef)

	if err != nil {
		return nil, errs.NewUnexpectedError("Error while querying UserCoins table " + err.Error())
	}

	return userCoins, nil
}

func (c *UserCoinRepositoryDb) FindByUserCoinId(ctx context.Context, userId, userCoinId string) (*UserCoin, error) {
	var err error
	var userCoin UserCoin

	query := "SELECT id, user_id, coin, coin_ref FROM user_coins WHERE user_id = ? AND id = ? LIMIT 1"

	err = c.client.GetContext(ctx, &userCoin, query, userId, userCoinId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewNotFoundError("The coin was not found")
		}

		return nil, errs.NewUnexpectedError("Error while querying UserCoins table " + err.Error())
	}

	return &userCoin, nil
}

func (c *UserCoinRepositoryDb) FindDifferentsCoins(ctx context.Context) ([]UserCoin, error) {
	var err error
	UserCoins := make([]UserCoin, 0)

	query := "SELECT DISTINCT coin, coin_ref FROM user_coins"

	err = c.client.SelectContext(ctx, &UserCoins, query)

	if err != nil {
		return nil, errs.NewUnexpectedError("Error while querying UserCoins table " + err.Error())
	}

	return UserCoins, nil
}

func (c *UserCoinRepositoryDb) Save(ctx context.Context, u UserCoin) (*UserCoin, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	query := "INSERT INTO user_coins (user_id, coin, coin_ref) values (?, ?, ?)"

	result, err := c.client.ExecContext(ctx, query, u.UserId, u.Coin, u.CoinRef)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			loggerWithCxt.Error("The user has already added that coin: " + err.Error())
			return nil, errs.NewConflictError("The user has already added that coin")
		}
		loggerWithCxt.Error("Error while creating new user coin: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	id, err := result.LastInsertId()
	if err != nil {
		loggerWithCxt.Error("Error while getting last insert id for new user coin: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}
	u.Id = int(id)
	return &u, nil
}

func (c *UserCoinRepositoryDb) Update(ctx context.Context, u UserCoin) (*UserCoin, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	query := "UPDATE user_coins SET coin = ?, coin_ref = ? WHERE id = ? AND user_id = ?"

	result, err := c.client.ExecContext(ctx, query, u.Coin, u.CoinRef, u.Id, u.UserId)
	if err != nil {
		loggerWithCxt.Error("Error while updating user coin: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		loggerWithCxt.Error("Error while updating user coin: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	if rowsAffected == 0 {
		loggerWithCxt.Error("item not found or user unauthorized to update")
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	return &u, nil
}

func (c *UserCoinRepositoryDb) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM user_coins WHERE id = ?"

	_, err := c.client.ExecContext(ctx, query, id)
	if err != nil {
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Error while deleting new user coin: " + err.Error())
		return errs.NewUnexpectedError("Unexpected error from database")
	}

	return nil
}
