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

type BrokerSite2 struct {
	client  *http.Client
	baseURL string
}

func NewBrokerSite2(client *http.Client, baseURL string) BrokerSite {
	return &BrokerSite2{client, baseURL}
}

func (d *BrokerSite2) GetName() string { return "CoinMarketCap" }

func (d *BrokerSite2) GetPrice(ctx context.Context, coin, coinRef string) (*float64, error) {
	loggerWithCxt := logger.NewLogger(ctx)
	coinSanitized := strings.ToUpper(coin)
	coinRefSanitized := strings.ToUpper(coinRef)

	baseUrl := fmt.Sprintf("%s/v1/tools/price-conversion", d.baseURL)
	queryParams := map[string]string{"symbol": coinSanitized, "amount": "1", "convert": coinRefSanitized}

	url := helpers.BuildUrl(baseUrl, queryParams)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		loggerWithCxt.Error("Error creating http request: " + err.Error())
		return nil, errs.NewUnexpectedError("Error creating http request")
	}
	key := os.Getenv("CMC_API_KEY")
	req.Header.Set("X-CMC_PRO_API_KEY", key)
	req.Header.Set("Accept", "application/json")
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

	data, err = helpers.GetAttribute(data, fmt.Sprintf(`data.quote.%s`, coinRefSanitized))

	if err != nil {
		loggerWithCxt.Error("Error parsing the response " + err.Error())
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	price, ok := data["price"]
	if !ok || price == nil {
		loggerWithCxt.Error(fmt.Sprintf("Error parsing the response or the price of %s %s is null", coin, coinRef))
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	value, ok := price.(float64)

	if !ok {
		loggerWithCxt.Error("Error parsing the response")
		return nil, errs.NewUnexpectedError("Error decoding response")
	}

	return &value, nil
}
