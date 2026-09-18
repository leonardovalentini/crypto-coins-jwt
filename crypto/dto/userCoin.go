package dto

import (
	"github.com/leonardovalentini/crypto-coins/lib/errs"
)

type UserCoinRequest struct {
	Coin    string `json:"coin" xml:"coin" example:"btc" minlength:"3" maxlength:"3"`
	CoinRef string `json:"coin_ref" xml:"coin_ref" example:"eth" minlength:"3" maxlength:"3"`
}

type UserCoinResponse struct {
	Id      int    `json:"id" xml:"id" example:"1"`
	Coin    string `json:"coin" xml:"coin" example:"btc" minlength:"3" maxlength:"3"`
	CoinRef string `json:"coin_ref" xml:"coin_ref" example:"eth" minlength:"3" maxlength:"3"`
}

func (r *UserCoinRequest) Validate() error {
	if len(r.Coin) != 3 {
		return errs.NewValidationError("The coin symbol must be three chars")
	}

	if len(r.CoinRef) != 3 {
		return errs.NewValidationError("The coin reference must be three chars")
	}

	return nil
}

type NewUserCoinResponse struct {
	Id int `json:"id" xml:"id"`
}
