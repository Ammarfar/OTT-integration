package app

import (
	"fmt"

	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/provider"
	"ott-integration/backend/internal/provider/netflix"
	"ott-integration/backend/internal/provider/netplay"
	"ott-integration/backend/internal/service"
	"ott-integration/backend/internal/store"
)

func NewService(cfg config.Config) (*service.Service, error) {
	activationStore, err := store.NewFileStore(cfg.StorePath)
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}

	providerRegistry := provider.NewRegistry(
		netplay.New(cfg.NetplayBaseURL, cfg.HTTPTimeout),
		netflix.New(),
	)
	return service.New(providerRegistry, activationStore, cfg), nil
}
