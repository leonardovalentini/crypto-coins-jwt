package domain

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type UserRepositoryDb struct {
	client *sqlx.DB
}

func NewUserRepositoryDb(client *sqlx.DB) UserRepository {
	return &UserRepositoryDb{client: client}
}

func (d *UserRepositoryDb) FindByUsername(ctx context.Context, username string) (*User, error) {
	var err error
	var user User

	query := "SELECT id, username, password FROM users WHERE username = ?"

	err = d.client.GetContext(ctx, &user, query, username)
	if err != nil {
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Error while creating new user: " + err.Error())
		if err == sql.ErrNoRows {
			return nil, errs.NewAuthenticationError("Invalid username or password")
		}
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	return &user, nil
}

func (d *UserRepositoryDb) Save(ctx context.Context, u User) (*User, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	query := "INSERT INTO users (username, password) values (?, ?)"

	result, err := d.client.ExecContext(ctx, query, u.Username, u.Password)
	if err != nil {
		loggerWithCxt.Error("Error while creating new user: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	id, err := result.LastInsertId()
	if err != nil {
		loggerWithCxt.Error("Error while getting last insert id for new user: " + err.Error())
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}
	u.Id = int(id)

	return &u, nil
}

func (d *UserRepositoryDb) CheckIfUsernameExists(ctx context.Context, username string) (*bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	err := d.client.GetContext(ctx, &exists, query, username)
	if err != nil {
		return nil, errs.NewUnexpectedError("Unexpected error from database")
	}

	return &exists, nil
}
