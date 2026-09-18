package domain

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/stretchr/testify/assert"
)

var price float64 = 79547

func setupBrokerSite1Tests(t *testing.T, coin, coinRef, url string, forceTransportLevel bool, httpRespCode int, httpRespBody string) (func(), BrokerSite) {
	var server *httptest.Server
	if url == "" {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if forceTransportLevel {
				hj, _ := w.(http.Hijacker)
				conn, _, _ := hj.Hijack()
				conn.Close()
				return
			}
			expectedUrl := "/api/v3/simple/price"
			assert.Equal(t, http.MethodGet, r.Method, fmt.Sprintf("Wrong path: got %v, want %v", r.Method, http.MethodGet))
			assert.Equal(t, expectedUrl, r.URL.Path, fmt.Sprintf("Wrong path: got %v, want %v", r.URL.Path, expectedUrl))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpRespCode)
			w.Write([]byte(httpRespBody))
		}))
	}

	var broker BrokerSite
	if url == "" {
		broker = NewBrokerSite1(server.Client(), server.URL)
	} else {
		broker = NewBrokerSite1(&http.Client{}, url)
	}

	return func() {
		if server != nil {
			server.Close()
		}
	}, broker
}

func Test_get_name(t *testing.T) {
	teardown, broker := setupBrokerSite1Tests(t, "", "", "", false, 200, "")
	defer teardown()

	name := broker.GetName()
	assert.Equal(t, "CoinGecko", name, fmt.Sprintf("Wrong error: got %v, want %v", name, "CoinGecko"))
}

func Test_get_price(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		coin                string
		coinRef             string
		forceTransportLevel bool
		httpRespCode        int
		httpRespBody        string
		expectedError       error
		expectedResponse    *float64
	}{
		{
			name:             "New request error",
			url:              "	", // unescaped horizontal-tab
			expectedResponse: nil,
			expectedError:    errs.NewUnexpectedError("Error creating http request"),
		},
		{
			name:                "Client do error",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: true,
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error creating http client"),
		},
		{
			name:                "Wrong http code",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusCreated,
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Wrong http code"),
		},
		{
			name:                "Wrong http body",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "Wrong body",
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error decoding response"),
		},
		{
			name:                "Wrong http body - coin",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "{\"eth\":{\"usd\":79547}}",
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error decoding response"),
		},
		{
			name:                "Wrong http body - coin invalid value",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "{\"btc\":\"price\"}",
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error decoding response"),
		},
		{
			name:                "Wrong http body - coin ref",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "{\"btc\":{\"eth\":79547}}",
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error decoding response"),
		},
		{
			name:                "Wrong http body - coin ref invalid value",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "{\"btc\":{\"usd\":\"price\"}}",
			expectedResponse:    nil,
			expectedError:       errs.NewUnexpectedError("Error decoding response"),
		},
		{
			name:                "Success",
			coin:                "btc",
			coinRef:             "usd",
			forceTransportLevel: false,
			httpRespCode:        http.StatusOK,
			httpRespBody:        "{\"btc\":{\"usd\":79547}}",
			expectedResponse:    &price,
			expectedError:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			teardown, broker := setupBrokerSite1Tests(t, tt.coin, tt.coinRef, tt.url, tt.forceTransportLevel, tt.httpRespCode, tt.httpRespBody)
			defer teardown()

			price, err := broker.GetPrice(ctx, tt.coin, tt.coinRef)

			assert.Equal(t, tt.expectedError, err, fmt.Sprintf("Wrong error: got %v, want %v", err, tt.expectedError))
			assert.Equal(t, tt.expectedResponse, price, fmt.Sprintf("Wrong error: got %v, want %v", price, tt.expectedResponse))
		})
	}

}
