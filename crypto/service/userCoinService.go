package service

import (
	"context"
	"strconv"

	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/lib/errs"
)

//go:generate mockgen -destination=../tests/mocks/service/mockUserCoinService.go -package=service github.com/leonardovalentini/crypto-coins/crypto/service UserCoinService
type UserCoinService interface {
	GetAllUserCoinsByUserId(ctx context.Context, id string) ([]domain.UserCoin, error)
	NewUserCoin(ctx context.Context, userId string, req dto.UserCoinRequest) (*dto.NewUserCoinResponse, error)
	UpdateUserCoin(ctx context.Context, userId, userCoinId string, req dto.UserCoinRequest) (*domain.UserCoin, error)
	DeleteUserCoin(ctx context.Context, userId, userCoinId string) error
}

type DefaultUserCoinService struct {
	repo domain.UserCoinRepository
}

func NewUserCoinService(repository domain.UserCoinRepository) UserCoinService {
	return &DefaultUserCoinService{repo: repository}
}

func (s *DefaultUserCoinService) GetAllUserCoinsByUserId(ctx context.Context, id string) ([]domain.UserCoin, error) {
	return s.repo.FindAllByUserId(ctx, id)
}

func (s *DefaultUserCoinService) NewUserCoin(ctx context.Context, userId string, req dto.UserCoinRequest) (*dto.NewUserCoinResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	userCoins, err := s.repo.FindBy(ctx, userId, req.Coin, req.CoinRef)
	if err != nil {
		return nil, err
	}
	if len(userCoins) != 0 {
		return nil, errs.NewConflictError("The user has already added that coin")
	}

	userCoin := domain.UserCoin{
		UserId:  userId,
		Coin:    req.Coin,
		CoinRef: req.CoinRef,
	}

	newUserCoin, err := s.repo.Save(ctx, userCoin)
	if err != nil {
		return nil, err
	}

	return newUserCoin.ToNewUserCoinResponseDto(), nil
}

func (s *DefaultUserCoinService) UpdateUserCoin(ctx context.Context, userId, userCoinId string, req dto.UserCoinRequest) (*domain.UserCoin, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	currentUserCoin, err := s.repo.FindByUserCoinId(ctx, userId, userCoinId)
	if err != nil {
		return nil, err
	}

	if currentUserCoin.Coin == req.Coin && currentUserCoin.CoinRef == req.CoinRef {
		return currentUserCoin, nil
	}

	existUserCoin, err := s.repo.FindBy(ctx, userId, req.Coin, req.CoinRef)
	if err != nil {
		return nil, err
	}
	if len(existUserCoin) != 0 {
		return nil, errs.NewConflictError("The user has already added that coin")
	}

	userCoin := domain.UserCoin{
		Id:      currentUserCoin.Id,
		UserId:  currentUserCoin.UserId,
		Coin:    req.Coin,
		CoinRef: req.CoinRef,
	}

	updatedUserCoin, err := s.repo.Update(ctx, userCoin)
	if err != nil {
		return nil, err
	}

	return updatedUserCoin, nil
}

func (s *DefaultUserCoinService) DeleteUserCoin(ctx context.Context, userId, userCoinId string) error {
	userCoin, err := s.repo.FindByUserCoinId(ctx, userId, userCoinId)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, strconv.Itoa(userCoin.Id))
	if err != nil {
		return err
	}

	return nil
}
