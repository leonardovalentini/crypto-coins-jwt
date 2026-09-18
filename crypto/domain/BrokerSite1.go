package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type BrokerSite1 struct {
	client  *http.Client
	baseURL string
}

func NewBrokerSite1(client *http.Client, baseURL string) BrokerSite {
	return &BrokerSite1{client, baseURL}
}
func (d *BrokerSite1) GetName() string { return "CoinGecko" }

func (d *BrokerSite1) GetPrice(ctx context.Context, coin, coinRef string) (*float64, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	coinSanitized := strings.ToLower(coin)
	coinRefSanitized := strings.ToLower(coinRef)

	baseUrl := fmt.Sprintf("%s/api/v3/simple/price", d.baseURL)
	queryParams := map[string]string{"vs_currencies": coinRefSanitized, "symbols": coinSanitized}

	url := helpers.BuildUrl(baseUrl, queryParams)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		loggerWithCxt.Error("Error creating http request: " + err.Error())
		return nil, errs.NewUnexpectedError("Error creating http request")
	}
	key := os.Getenv("GECKO_API_KEY")
	req.Header.Set("x-cg-demo-api-key", key)
	req = req.WithContext(ctx)
	resp, err := d.client.Do(req)
	if err != nil {
		loggerWithCxt.Error("Error creating http client: " + err.Error())
		return nil, errs.NewUnexpectedError("Error creating http client")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		loggerWithCxt.Error("Wrong http code")
		return nil, errs.NewUnexpectedError("Wrong http code")
	}

	var data map[string]any
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		loggerWithCxt.Error("Error parsing the response " + err.Error())
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	coinData, err := helpers.GetAttribute(data, coinSanitized)
	if err != nil {
		loggerWithCxt.Error("Error parsing the response " + err.Error())
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	price, ok := coinData[coinRefSanitized]
	if !ok {
		loggerWithCxt.Error("Error parsing the response")
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	value, ok := price.(float64)

	if !ok {
		loggerWithCxt.Error("Error parsing the response")
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	return &value, nil
}
