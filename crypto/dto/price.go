package dto

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/leonardovalentini/crypto-coins/lib/errs"
)

type EmptyResponse struct{}

type GetPricesRequest struct {
	Page      int
	Size      int
	Coin      *string
	CoinRef   *string
	Price     *float64
	StartDate *time.Time
	EndDate   *time.Time
	Site      *string
}

var validSymbolRegex = regexp.MustCompile(`^[A-Z]{3}$`)

func NewGetPricesRequest(r *http.Request) (*GetPricesRequest, error) {
	queryUrl := r.URL.Query()
	page, err := strconv.Atoi(queryUrl.Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	size, err := strconv.Atoi(queryUrl.Get("size"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}
	response := GetPricesRequest{
		Page: page,
		Size: size,
	}

	if queryUrl.Has("coin") {
		coin := queryUrl.Get("coin")
		response.Coin = &coin
	}

	if queryUrl.Has("coin_ref") {
		coinRef := queryUrl.Get("coin_ref")
		response.CoinRef = &coinRef
	}

	if queryUrl.Has("price") {
		price, err := strconv.ParseFloat(queryUrl.Get("price"), 64)
		if err != nil {
			return nil, errs.NewValidationError("Invalid price value")
		}
		response.Price = &price
	}

	if queryUrl.Has("start_date") {
		startDate, err := time.Parse(time.RFC3339, queryUrl.Get("start_date"))
		if err != nil {
			return nil, errs.NewValidationError("Invalid start date value, should be in ISO format")
		}
		response.StartDate = &startDate
	}

	if queryUrl.Has("end_date") {
		endDate, err := time.Parse(time.RFC3339, queryUrl.Get("end_date"))
		if err != nil {
			return nil, errs.NewValidationError("Invalid end date value, should be in ISO format")
		}
		response.EndDate = &endDate
	}

	if queryUrl.Has("site") {
		site := queryUrl.Get("site")
		response.Site = &site
	}

	return &response, nil
}

func (pr *GetPricesRequest) Validate() error {
	if pr.Coin != nil && !validSymbolRegex.MatchString(*pr.Coin) {
		return errs.NewValidationError("The coin symbol must be three chars")
	}

	if pr.CoinRef != nil && !validSymbolRegex.MatchString(*pr.CoinRef) {
		return errs.NewValidationError("The coin reference symbol must be three chars")
	}

	if pr.Price != nil && (math.IsNaN(*pr.Price) || *pr.Price < 0) {
		return errs.NewValidationError("The price shold be a positive number")
	}

	if pr.StartDate != nil && time.Now().Before(*pr.StartDate) {
		return errs.NewValidationError("The start date should be before today")
	}

	if pr.EndDate != nil && time.Now().Before(*pr.EndDate) {
		return errs.NewValidationError("The end date should be before today")
	}

	if pr.StartDate != nil && pr.EndDate != nil && pr.EndDate.Before(*pr.StartDate) {
		return errs.NewValidationError("The start date should be before end date")
	}
	return nil
}

type PriceResponse struct {
	Id      int       `json:"-" xml:"-"`
	Coin    string    `json:"coin" xml:"coin" example:"btc" minlength:"3" maxlength:"3"`
	CoinRef string    `json:"coin_ref" xml:"coin_ref" example:"eth" minlength:"3" maxlength:"3"`
	Date    time.Time `json:"date" xml:"date" example:"2026-08-19T18:33:34Z"`
	Price   float64   `json:"price" xml:"price" example:"12.34"`
	Site    string    `json:"site" xml:"site" example:"CoinGecko"`
}

type SummaryResponse struct {
	Coin    map[string]int `json:"coin,omitempty" xml:"coin,omitempty" swaggertype:"object,integer" example:"btc:2,eth:3"`
	CoinRef map[string]int `json:"coin_ref,omitempty" xml:"coin_ref,omitempty" swaggertype:"object,integer" example:"ldc:1,usd:4"`
	Site    map[string]int `json:"site,omitempty" xml:"site,omitempty" swaggertype:"object,integer" example:"CoinGecko:4,CoinMarketCap:1"`
}

type PricesResponse struct {
	Page       int             `json:"page" xml:"page" example:"0"`
	Size       int             `json:"size" xml:"size" example:"10"`
	TotalItems int             `json:"total_items" xml:"total_items" example:"1"`
	TotalPages int             `json:"total_pages" xml:"total_pages" example:"1"`
	Items      []PriceResponse `json:"items" xml:"items"`
	Summary    SummaryResponse `json:"summary" xml:"summary"`
}
