package domain

import "context"

//go:generate mockgen -destination=../tests/mocks/domain/mockBrokerSiteRepository.go -package=domain github.com/leonardovalentini/crypto-coins/crypto/domain BrokerSite
type BrokerSite interface {
	GetName() string
	GetPrice(ctx context.Context, coin, coinRef string) (*float64, error)
}
