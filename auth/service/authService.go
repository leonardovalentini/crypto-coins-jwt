package service

import (
	"context"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/leonardovalentini/crypto-coins/auth/domain"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	sessionStore = make(map[string]string)
	sessionMutex sync.RWMutex
)

//go:generate mockgen -destination=../tests/mocks/service/mockAuthService.go -package=service github.com/leonardovalentini/crypto-coins/auth/service AuthService
type AuthService interface {
	Login(ctx context.Context, username, password string) (*http.Cookie, error)
	Logout(cookie *http.Cookie)
	Register(ctx context.Context, username, password string) (*int, error)
	GetUserId(ctx context.Context, cookie *http.Cookie) (*string, error)
}

type DefaultAuthService struct {
	repo domain.UserRepository
	rg   func(b []byte) (n int, err error)
	now  func() time.Time
}

func NewAuthService(repository domain.UserRepository, randomGenerator func(b []byte) (int, error), n func() time.Time) AuthService {
	return &DefaultAuthService{repo: repository, rg: randomGenerator, now: n}
}

func (s *DefaultAuthService) Login(ctx context.Context, username, password string) (*http.Cookie, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errs.NewAuthenticationError("Invalid username or password")
	}

	sessionToken, err := s.generateToken()
	if err != nil {
		return nil, errs.NewUnexpectedError("Unexpected error")
	}
	sessionMutex.Lock()
	sessionStore[*sessionToken] = strconv.Itoa(user.Id)
	sessionMutex.Unlock()

	return &http.Cookie{
		Name:     "session_token",
		Value:    *sessionToken,
		Path:     "/",
		Expires:  s.now().Add(30 * time.Minute),
		MaxAge:   1800,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil

}

func (s *DefaultAuthService) Register(ctx context.Context, username, password string) (*int, error) {
	exists, err := s.repo.CheckIfUsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}

	if *exists {
		return nil, errs.NewConflictError("The user is already registered.")
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewUnexpectedError("An error occurred while encrypting the password.")
	}

	newUser := domain.User{
		Username: username,
		Password: string(hashPass),
	}

	userSaved, err := s.repo.Save(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &userSaved.Id, nil
}

func (s *DefaultAuthService) Logout(cookie *http.Cookie) {
	sessionMutex.Lock()
	delete(sessionStore, cookie.Value)
	sessionMutex.Unlock()
}

func (s *DefaultAuthService) GetUserId(ctx context.Context, cookie *http.Cookie) (*string, error) {
	sessionMutex.RLock()
	userId, exists := sessionStore[cookie.Value]
	sessionMutex.RUnlock()
	if !exists {
		loggerWithCxt := logger.NewLogger(ctx)
		loggerWithCxt.Error("Unauthorized: no session found")
		return nil, errs.NewAuthenticationError("Unauthorized: No session found")
	}

	return &userId, nil
}

func (s *DefaultAuthService) generateToken() (*string, error) {
	b := make([]byte, 32)
	if _, err := s.rg(b); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(b)
	return &token, nil
}
