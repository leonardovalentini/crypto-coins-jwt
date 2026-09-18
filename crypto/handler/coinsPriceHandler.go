package handler

import (
	"context"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/crypto/service"
	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

type CoinsPriceHandler struct {
	service       service.CoinsPriceService
	jobPricesBusy chan string
}

func NewCoinsPriceHandler(service service.CoinsPriceService) *CoinsPriceHandler {
	ch := make(chan string, 1)
	ch <- "Free"

	return &CoinsPriceHandler{service: service, jobPricesBusy: ch}
}

func (h *CoinsPriceHandler) RegisterPublicRoutes(router *mux.Router) {
	pricesCoinRouter := router.PathPrefix("/broker-prices").Subrouter()

	pricesCoinRouter.
		HandleFunc("", h.GetBrokerPrices).
		Methods(http.MethodGet).
		Name("JobPrices")
}

func (h *CoinsPriceHandler) RegisterPrivateRoutes(router *mux.Router) {
	pricesCoinRouter := router.PathPrefix("/prices").Subrouter()

	pricesCoinRouter.
		HandleFunc("", h.GetAllPrices).
		Methods(http.MethodGet).
		Name("GetAllPrices")
}

// GetBrokerPrices godoc
//
//	@Summary		Job to complete prices
//	@Description	Job to complete current currency prices from brokers.
//	@Tags			prices
//	@Accept			json
//	@Produce		json,xml
//	@Failure		503	{object}	helpers.HttpError	"Job is already running"
//	@Success		202	{object}	dto.EmptyResponse	"Accepted"
//	@Router			/jobs/broker-prices [get]
func (h *CoinsPriceHandler) GetBrokerPrices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-h.jobPricesBusy:
	default:
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Job is already running")
		err := errs.NewServiceUnavailableError("Job is already running")
		helpers.HandleError(w, r, err)
		return
	}
	defer func() {
		h.jobPricesBusy <- "Free"
	}()

	jobID := uuid.New().String()
	jobCtx := contextKey.WithJobId(context.Background(), jobID)

	helpers.WriteResponse(w, r, http.StatusAccepted, dto.EmptyResponse{})
	h.service.JobPrices(jobCtx)
}

// GetAllPrices godoc
//
//	@Summary		Get historical price lists
//	@Description	Retrieve filtered and paginated historical price lists.
//	@Tags			prices
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Param			page		query		int		false	"Page number starting 0"	default(0)
//	@Param			size		query		int		false	"Page size"					default(10)
//	@Param			coin		query		string	false	"3-character code"			minlength(3)		maxlength(3)	default()	example(btc)
//	@Param			coin_ref	query		string	false	"3-character code"			minlength(3)		maxlength(3)	default()	example(eth)
//	@Param			price		query		number	false	"Price"						minimum(0)			default()		example(12.34)
//	@Param			start_date	query		string	false	"Start date (ISO-8601)"		format(date-time)	default()		example(2026-08-21T15:30:00Z)
//	@Param			end_date	query		string	false	"End date (ISO-8601)"		format(date-time)	default()		example(2026-08-21T15:30:00Z)
//	@Param			site		query		string	false	"Site"						default()			example(CoinGecko)
//	@Success		200			{object}	dto.PricesResponse
//	@Failure		401			{object}	helpers.HttpError	"User id not found"
//	@Failure		422			{object}	helpers.HttpError	"Invalid price value<br>Invalid start date value, should be in ISO format<br>Invalid end date value, should be in ISO format<br>The coin symbol must be three chars<br>The coin reference must be three chars<br>The coin symbol must be three chars<br>The coin reference must be three chars<br>The price shold be a positive number<br>The start date should be before today<br>The end date should be before today<br>The start date should be before end date"
//	@Failure		500			{object}	helpers.HttpError	"Internal Server Error<br>Unexpected error from database
//	@Router			/api/prices [get]
func (h *CoinsPriceHandler) GetAllPrices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	loggerWithCxt := logger.NewLogger(ctx)
	userId, err := contextKey.GetUserId(ctx)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	gpr, err := dto.NewGetPricesRequest(r)
	if err != nil {
		loggerWithCxt.Error("Error parsing get prices request: " + err.Error())
		helpers.HandleError(w, r, err)
		return
	}
	err = gpr.Validate()
	if err != nil {
		loggerWithCxt.Error("Error validating get prices request: " + err.Error())
		helpers.HandleError(w, r, err)
		return
	}

	prices, total, summary, err := h.service.GetPrices(ctx, *userId, gpr)
	if err != nil {
		loggerWithCxt.Error("Error in GetPrices: " + err.Error())
		helpers.HandleError(w, r, err)
		return
	}

	items := []dto.PriceResponse{}
	for _, price := range prices {
		items = append(items, *price.ToDto())
	}

	totalPages := int(math.Ceil(float64(*total) / float64(gpr.Size)))

	response := dto.PricesResponse{
		Page:       gpr.Page,
		Size:       gpr.Size,
		Items:      items,
		TotalItems: *total,
		TotalPages: totalPages,
		Summary:    *summary.ToDto(),
	}
	helpers.WriteResponse(w, r, http.StatusOK, &response)
}
