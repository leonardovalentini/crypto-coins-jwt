package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/crypto/service"
	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
)

type UserCoinHandler struct {
	service service.UserCoinService
}

func NewUserCoinHandler(service service.UserCoinService) *UserCoinHandler {
	return &UserCoinHandler{service: service}
}

func (h *UserCoinHandler) RegisterPrivateRoutes(router *mux.Router) {
	userCoinRouter := router.PathPrefix("/user-coins").Subrouter()

	userCoinRouter.
		HandleFunc("", h.GetAllUserCoins).
		Methods(http.MethodGet).
		Name("GetAllUserCoins")
	userCoinRouter.
		HandleFunc("", h.NewUserCoin).
		Methods(http.MethodPost).
		Name("NewUserCoin")
	userCoinRouter.
		HandleFunc("/{user_coin_id:[0-9]+}", h.UpdateUserCoin).
		Methods(http.MethodPut).
		Name("NewUserCoin")
	userCoinRouter.
		HandleFunc("/{user_coin_id:[0-9]+}", h.DeleteUserCoin).
		Methods(http.MethodDelete).
		Name("DeleteUserCoin")
}

// GetAllUserCoins godoc
//
//	@Summary		Get user coins
//	@Description	returns all the coins configured by the user
//	@Tags			user-coins
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Success		200	{array}		dto.UserCoinResponse
//	@Failure		401	{object}	helpers.HttpError	"User id not found"
//	@Failure		500	{object}	helpers.HttpError
//	@Router			/api/user-coins [get]
func (h *UserCoinHandler) GetAllUserCoins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, err := contextKey.GetUserId(ctx)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	UserCoins, err := h.service.GetAllUserCoinsByUserId(ctx, *userId)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}
	userCoinsResponse := []dto.UserCoinResponse{}

	for _, userCoin := range UserCoins {
		userCoinsResponse = append(userCoinsResponse, *userCoin.ToDto())
	}

	helpers.WriteResponse(w, r, http.StatusOK, userCoinsResponse)
}

// NewUserCoin godoc
//
//	@Summary		Add user coin
//	@Description	Adds a new currency configuration to the user's list.
//	@Tags			user-coins
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Param			body	body		dto.UserCoinRequest	true	"Currency info"
//	@Success		200		{object}	dto.NewUserCoinResponse
//	@Failure		400		{object}	helpers.HttpError	"Invalid request body<br>The coin symbol must be three chars<br>The coin reference must be three chars"
//	@Failure		401		{object}	helpers.HttpError	"User id not found"
//	@Failure		409		{object}	helpers.HttpError	"The user has already added that coin"
//	@Failure		422		{object}	helpers.HttpError	"The coin symbol must be three chars<br>The coin reference must be three chars"
//	@Failure		500		{object}	helpers.HttpError
//	@Router			/api/user-coins [post]
func (h *UserCoinHandler) NewUserCoin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, err := contextKey.GetUserId(ctx)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	var request dto.UserCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.HandleError(w, r, errs.NewBadRequestError("Invalid request body"))
		return
	}

	newUserCoin, err := h.service.NewUserCoin(ctx, *userId, request)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	helpers.WriteResponse(w, r, http.StatusCreated, newUserCoin)
}

// UpdateUserCoin godoc
//
//	@Summary		Update user coin
//	@Description	Update currency configuration to the user's list.
//	@Tags			user-coins
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Param			id		path		int					true	"Currency ID"
//	@Param			body	body		dto.UserCoinRequest	true	"Currency info"
//	@Success		200		{object}	dto.UserCoinResponse
//	@Failure		400		{object}	helpers.HttpError	"Invalid request body<br>The coin symbol must be three chars<br>The coin reference must be three chars"
//	@Failure		401		{object}	helpers.HttpError	"User id not found"
//	@Failure		404		{object}	helpers.HttpError	"The coin was not found"
//	@Failure		409		{object}	helpers.HttpError	"The user has already added that coin"
//	@Failure		422		{object}	helpers.HttpError	"The coin symbol must be three chars<br>The coin reference must be three chars"
//	@Failure		500		{object}	helpers.HttpError
//	@Router			/api/user-coins/{id} [put]
func (h *UserCoinHandler) UpdateUserCoin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, err := contextKey.GetUserId(ctx)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	vars := mux.Vars(r)
	userCoinId := vars["user_coin_id"]

	var request dto.UserCoinRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.HandleError(w, r, errs.NewBadRequestError("Invalid request body"))
		return
	}

	updatedUserCoin, err := h.service.UpdateUserCoin(ctx, *userId, userCoinId, request)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	userCoinResponse := updatedUserCoin.ToDto()

	helpers.WriteResponse(w, r, http.StatusOK, userCoinResponse)
}

// DeleteUserCoin godoc
//
//	@Summary		Delete user coin
//	@Description	Delete currency configuration to the user's list.
//	@Tags			user-coins
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Param			id	path		int	true	"Currency ID"
//	@Success		200	{object}	dto.EmptyResponse
//	@Failure		401	{object}	helpers.HttpError	"User id not found"
//	@Failure		404	{object}	helpers.HttpError	"The coin was not found"
//	@Failure		500	{object}	helpers.HttpError
//	@Router			/api/user-coins/{id} [delete]
func (h *UserCoinHandler) DeleteUserCoin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, err := contextKey.GetUserId(ctx)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	vars := mux.Vars(r)
	userCoinId := vars["user_coin_id"]

	err = h.service.DeleteUserCoin(ctx, *userId, userCoinId)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	helpers.WriteResponse(w, r, http.StatusOK, dto.EmptyResponse{})
}
