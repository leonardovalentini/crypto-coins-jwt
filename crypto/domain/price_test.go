package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/stretchr/testify/assert"
)

func Test_price_to_dto(t *testing.T) {
	p := Price{
		Id:      1,
		Coin:    "btc",
		CoinRef: "usd",
		Date:    time.Now(),
		Price:   0.1,
		Site:    "CoinGecko",
	}
	response := p.ToDto()
	expectedResponse := &dto.PriceResponse{
		Id:      1,
		Coin:    "btc",
		CoinRef: "usd",
		Date:    p.Date,
		Price:   0.1,
		Site:    "CoinGecko",
	}
	assert.Equal(t, expectedResponse, response, fmt.Sprintf("Wrong response: got %v, want %v", response, expectedResponse))
}

func Test_summary_to_dto(t *testing.T) {
	s := Summary{
		Coin:    map[string]int{"btc": 1},
		CoinRef: map[string]int{"usd": 1},
		Site:    map[string]int{"CoinGecko": 1},
	}
	response := s.ToDto()
	expectedResponse := &dto.SummaryResponse{
		Coin:    map[string]int{"btc": 1},
		CoinRef: map[string]int{"usd": 1},
		Site:    map[string]int{"CoinGecko": 1},
	}
	assert.Equal(t, expectedResponse, response, fmt.Sprintf("Wrong response: got %v, want %v", response, expectedResponse))
}
