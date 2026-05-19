package provider

import (
	"context"

	"ott-integration/backend/internal/domain"
)

type Client interface {
	Name() domain.ProviderName
	Plans() []domain.Plan
	Subscribe(ctx context.Context, cmd domain.SubscribeCommand, idempotencyKey string) (domain.ProviderSubscribeResult, error)
	Activate(ctx context.Context, activationToken string) (domain.ProviderActivateResult, error)
	GetSubscriptionStatus(ctx context.Context, activationToken string) (domain.ProviderStatusResult, error)
}

type Registry struct {
	clients map[domain.ProviderName]Client
}

func NewRegistry(clients ...Client) *Registry {
	registry := &Registry{clients: make(map[domain.ProviderName]Client, len(clients))}
	for _, client := range clients {
		registry.clients[client.Name()] = client
	}
	return registry
}

func (r *Registry) Get(name domain.ProviderName) (Client, bool) {
	client, ok := r.clients[name]
	return client, ok
}

func (r *Registry) List() []domain.ProviderCapability {
	result := make([]domain.ProviderCapability, 0, len(r.clients))
	for _, client := range r.clients {
		result = append(result, domain.ProviderCapability{
			Name:  client.Name(),
			Plans: client.Plans(),
		})
	}
	return result
}
