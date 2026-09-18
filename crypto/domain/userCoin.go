package domain

import (
	"context"

	"github.com/leonardovalentini/crypto-coins/crypto/dto"
)

type UserCoin struct {
	Id      int    `json:"id" xml:"id" db:"id"`
	UserId  string `json:"-" xml:"" db:"user_id"`
	Coin    string `json:"coin" xml:"coin" db:"coin"`
	CoinRef string `json:"coin_ref" xml:"coin_ref" db:"coin_ref"`
}

func (c *UserCoin) ToNewUserCoinResponseDto() *dto.NewUserCoinResponse {
	return &dto.NewUserCoinResponse{
		Id: c.Id,
	}
}

func (c *UserCoin) ToDto() *dto.UserCoinResponse {
	return &dto.UserCoinResponse{
		Id:      c.Id,
		Coin:    c.Coin,
		CoinRef: c.CoinRef,
	}
}

//go:generate mockgen -destination=../tests/mocks/domain/mockUserCoinRepository.go -package=domain github.com/leonardovalentini/crypto-coins/crypto/domain UserCoinRepository
type UserCoinRepository interface {
	FindAllByUserId(ctx context.Context, id string) ([]UserCoin, error)
	FindBy(ctx context.Context, userId, coin, coinRef string) ([]UserCoin, error)
	FindByUserCoinId(ctx context.Context, userId, userCoinId string) (*UserCoin, error)
	FindDifferentsCoins(ctx context.Context) ([]UserCoin, error)
	Save(ctx context.Context, u UserCoin) (*UserCoin, error)
	Update(ctx context.Context, u UserCoin) (*UserCoin, error)
	Delete(ctx context.Context, id string) error
}
