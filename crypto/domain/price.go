package domain

import (
	"context"
	"time"

	"github.com/leonardovalentini/crypto-coins/crypto/dto"
)

type Summary struct {
	Coin    map[string]int
	CoinRef map[string]int
	Site    map[string]int
}

func (d *Summary) ToDto() *dto.SummaryResponse {
	return &dto.SummaryResponse{
		Coin:    d.Coin,
		CoinRef: d.CoinRef,
		Site:    d.Site,
	}
}

type Price struct {
	Id      int       `json:"id" xml:"id" db:"id"`
	Coin    string    `json:"coin" xml:"coin" db:"coin"`
	CoinRef string    `json:"coin_ref" xml:"coin_ref" db:"coin_ref"`
	Date    time.Time `json:"date" xml:"date" db:"date"`
	Price   float64   `json:"price" xml:"price" db:"price"`
	Site    string    `json:"site" xml:"site" db:"site"`
}

func (d *Price) ToDto() *dto.PriceResponse {
	return &dto.PriceResponse{
		Id:      d.Id,
		Coin:    d.Coin,
		CoinRef: d.CoinRef,
		Date:    d.Date,
		Price:   d.Price,
		Site:    d.Site,
	}
}

//go:generate mockgen -destination=../tests/mocks/domain/mockPriceRepository.go -package=domain github.com/leonardovalentini/crypto-coins/crypto/domain PriceRepository
type PriceRepository interface {
	FindAllBy(ctx context.Context, userId string, gpr *dto.GetPricesRequest) ([]Price, error)
	GetSummary(ctx context.Context, userId string, gpr *dto.GetPricesRequest) (*int, *Summary, error)
	Save(ctx context.Context, p Price) (*Price, error)
}

type SummaryRow struct {
	ColumnName string `json:"column_name" xml:"column_name" db:"column_name"`
	Value      string `json:"value" xml:"value" db:"value"`
	Count      int    `json:"count" xml:"count" db:"count"`
}
