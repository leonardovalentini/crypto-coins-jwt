package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leonardovalentini/crypto-coins/auth/dto"
	"github.com/leonardovalentini/crypto-coins/auth/service"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	authRouter := router.PathPrefix("/auth").Subrouter()

	authRouter.
		HandleFunc("/login", h.Login).
		Methods(http.MethodPost).
		Name("Login")
	authRouter.
		HandleFunc("/register", h.Register).
		Methods(http.MethodPost).
		Name("Register")
	authRouter.
		HandleFunc("/verify", h.Verify).
		Methods(http.MethodGet).
		Name("Verify")
	authRouter.
		HandleFunc("/logout", h.Logout).
		Methods(http.MethodPost).
		Name("Logout")
}

// Login godoc
//
//	@Summary		Log in and set auth cookie
//	@Description	Logs in a user using form data and sets an auth cookie
//	@Tags			auth
//	@Accept			multipart/form-data
//	@Produce		json,xml
//	@Param			username	formData	string	true	"username"
//	@Param			password	formData	string	true	"password"
//	@Success		200			{object}	dto.Response
//	@Header			200			{string}	Set-Cookie			"session_token=abcde12345; Path=/; HttpOnly; Max-Age=1800"
//	@Failure		401			{object}	helpers.HttpError	"Invalid username or password"
//	@Failure		500			{object}	helpers.HttpError	"Unexpected error from database"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.FormValue("username")
	password := r.FormValue("password")

	cookie, err := h.service.Login(ctx, username, password)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	http.SetCookie(w, cookie)
	helpers.WriteResponse(w, r, http.StatusOK, dto.Response{Message: "OK"})
}

// Register godoc
//
//	@Summary		Register new user
//	@Description	Register with username and password form fields
//	@Tags			auth
//	@Accept			multipart/form-data
//	@Produce		json,xml
//	@Param			username	formData	string	true	"username"
//	@Param			password	formData	string	true	"password"
//	@Success		201			{object}	dto.ResponseCreated
//	@Failure		401			{object}	helpers.HttpError	"Invalid username or password"
//	@Failure		409			{object}	helpers.HttpError	"The user is already registered."
//	@Failure		500			{object}	helpers.HttpError	"Unexpected error from database<br>An error occurred while encrypting the password."
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.FormValue("username")
	password := r.FormValue("password")

	userId, err := h.service.Register(ctx, username, password)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	helpers.WriteResponse(w, r, http.StatusCreated, dto.ResponseCreated{Id: *userId})
}

// Verify godoc
//
//	@Summary		Verify Auth cookie
//	@Description	Verify Auth cookie and return user id
//	@Tags			auth
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Success		200	{object}	dto.VerifyResponse
//	@Failure		401	{object}	helpers.HttpError	"Session cookie not found<br>Unauthorized: No session found"
//	@Failure		500	{object}	helpers.HttpError
//	@Router			/auth/verify [get]
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := getAuthCookie(r)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	id, err := h.service.GetUserId(ctx, cookie)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	v := dto.VerifyResponse{Id: *id}
	helpers.WriteResponse(w, r, http.StatusOK, &v)
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Clear auth cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json,xml
//	@Security		CookieAuth
//	@Success		200	{object}	dto.Response
//	@Failure		401	{object}	helpers.HttpError	"Session cookie not found"
//	@Failure		500	{object}	helpers.HttpError
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := getAuthCookie(r)
	if err != nil {
		helpers.HandleError(w, r, err)
		return
	}

	h.service.Logout(cookie)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	helpers.WriteResponse(w, r, http.StatusOK, dto.Response{Message: "OK"})
}

func getAuthCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		newAuthErr := errs.NewAuthenticationError("Session cookie not found")
		return nil, newAuthErr
	}

	return cookie, nil
}
