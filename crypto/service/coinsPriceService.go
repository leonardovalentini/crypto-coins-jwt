package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/dto"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
)

//go:generate mockgen -destination=../tests/mocks/service/mockCoinsPriceService.go -package=service github.com/leonardovalentini/crypto-coins/crypto/service CoinsPriceService
type CoinsPriceService interface {
	GetPrices(ctx context.Context, userId string, gpr *dto.GetPricesRequest) ([]domain.Price, *int, *domain.Summary, error)
	JobPrices(ctx context.Context)
}

type DefaultCoinsPriceService struct {
	ucr    domain.UserCoinRepository
	pr     domain.PriceRepository
	bs     []domain.BrokerSite
	logger logger.LoggerI
	now    func() time.Time
}

func NewCoinsPriceService(ucr domain.UserCoinRepository, pr domain.PriceRepository, bs []domain.BrokerSite, l logger.LoggerI, n func() time.Time) CoinsPriceService {
	return &DefaultCoinsPriceService{ucr: ucr, pr: pr, bs: bs, logger: l, now: n}
}

func (s *DefaultCoinsPriceService) GetPrices(ctx context.Context, userId string, gpr *dto.GetPricesRequest) ([]domain.Price, *int, *domain.Summary, error) {
	prices, err := s.pr.FindAllBy(ctx, userId, gpr)
	if err != nil {
		return nil, nil, nil, err
	}
	total, summary, err := s.pr.GetSummary(ctx, userId, gpr)
	if err != nil {
		return nil, nil, nil, err
	}
	return prices, total, summary, nil
}

func (s *DefaultCoinsPriceService) JobPrices(ctx context.Context) {
	s.logger.SetCtx(ctx)
	defer func() {
		if err := recover(); err != nil {
			s.logger.Error(fmt.Sprintf("Encountered unexpected error: %v", err))
		}
	}()

	coins, err := s.ucr.FindDifferentsCoins(ctx)
	if err != nil {
		s.logger.Error("Error in JobPrices: " + err.Error())
		return
	}

	var wg sync.WaitGroup
	for _, site := range s.bs {
		wg.Add(1)
		go func(ctx context.Context, site domain.BrokerSite, coins []domain.UserCoin) {
			defer wg.Done()

			s.getPricesBySite(ctx, site, coins)
		}(ctx, site, coins)
	}
	wg.Wait()
}

func (s *DefaultCoinsPriceService) getPricesBySite(ctx context.Context, bs domain.BrokerSite, coins []domain.UserCoin) {
	const maxConcurrency = 3

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, coin := range coins {
		wg.Add(1)
		sem <- struct{}{}

		go func(ctx context.Context, bs domain.BrokerSite, coin domain.UserCoin) {
			defer wg.Done()
			defer func() { <-sem }()

			s.getPricesBySiteAndCoin(ctx, bs, coin)
		}(ctx, bs, coin)

	}
	wg.Wait()
}

func (s *DefaultCoinsPriceService) getPricesBySiteAndCoin(ctx context.Context, bs domain.BrokerSite, coin domain.UserCoin) {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Error(fmt.Sprintf("Encountered unexpected error: %v", err))
		}
	}()

	price, err := bs.GetPrice(ctx, coin.Coin, coin.CoinRef)
	if err != nil {
		s.logger.Error("JobPrices: Unexpected error getting the price " + err.Error())
		return
	}

	p := domain.Price{
		Coin:    coin.Coin,
		CoinRef: coin.CoinRef,
		Date:    s.now(),
		Price:   *price,
		Site:    bs.GetName(),
	}

	_, err = s.pr.Save(ctx, p)
	if err != nil {
		s.logger.Error("JobPrices: Unexpected error saving the price: " + err.Error())
		return
	}
	s.logger.Info(fmt.Sprintf("JobPrices: Price saved successfully for %s/%s from %s", coin.Coin, coin.CoinRef, bs.GetName()))
}
