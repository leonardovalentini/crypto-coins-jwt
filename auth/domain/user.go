package domain

import "context"

type User struct {
	Id       int    `json:"id" xml:"id" db:"id"`
	Username string `json:"username" xml:"username" db:"username"`
	Password string `json:"password" xml:"password" db:"password"`
}

//go:generate mockgen -destination=../tests/mocks/domain/mockUserRepository.go -package=domain github.com/leonardovalentini/crypto-coins/auth/domain UserRepository
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	Save(ctx context.Context, u User) (*User, error)
	CheckIfUsernameExists(ctx context.Context, username string) (*bool, error)
}
