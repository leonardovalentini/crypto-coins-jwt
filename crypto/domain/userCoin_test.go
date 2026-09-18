package domain

import (
	"fmt"
	"testing"

	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/stretchr/testify/assert"
)

func Test_to_new_user_coin_response_dto(t *testing.T) {
	u := UserCoin{
		Id:      1,
		UserId:  "1",
		Coin:    "btc",
		CoinRef: "usd",
	}
	response := u.ToNewUserCoinResponseDto()
	expectedResponse := &dto.NewUserCoinResponse{Id: 1}
	assert.Equal(t, expectedResponse, response, fmt.Sprintf("Wrong response: got %v, want %v", response, expectedResponse))
}

func Test_user_coin_to_dto(t *testing.T) {
	u := UserCoin{
		Id:      1,
		UserId:  "1",
		Coin:    "btc",
		CoinRef: "usd",
	}
	response := u.ToDto()
	expectedResponse := &dto.UserCoinResponse{Id: 1, Coin: "btc", CoinRef: "usd"}
	assert.Equal(t, expectedResponse, response, fmt.Sprintf("Wrong response: got %v, want %v", response, expectedResponse))
}
