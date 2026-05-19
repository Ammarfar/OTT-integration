package service

import (
	"context"
	"testing"
	"time"

	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/domain"
	"ott-integration/backend/internal/provider"
)

type providerStub struct {
	subscribeResp domain.ProviderSubscribeResult
	activateResp  domain.ProviderActivateResult
	statusResp    domain.ProviderStatusResult
}

func (p providerStub) Name() domain.ProviderName {
	return domain.ProviderNetplay
}

func (p providerStub) Plans() []domain.Plan {
	return []domain.Plan{domain.PlanPremium30D}
}

func (p providerStub) Subscribe(_ context.Context, _ domain.SubscribeCommand, _ string) (domain.ProviderSubscribeResult, error) {
	return p.subscribeResp, nil
}

func (p providerStub) Activate(_ context.Context, _ string) (domain.ProviderActivateResult, error) {
	return p.activateResp, nil
}

func (p providerStub) GetSubscriptionStatus(_ context.Context, _ string) (domain.ProviderStatusResult, error) {
	return p.statusResp, nil
}

type memoryStore struct {
	record domain.SubscriptionRecord
}

func (m *memoryStore) Save(record domain.SubscriptionRecord) error {
	m.record = record
	return nil
}

func (m *memoryStore) GetByCode(code string) (domain.SubscriptionRecord, error) {
	if m.record.ActivationCode != code {
		return domain.SubscriptionRecord{}, ErrActivationNotFound
	}
	return m.record, nil
}

func (m *memoryStore) Update(record domain.SubscriptionRecord) error {
	m.record = record
	return nil
}

func TestSubscribeCreatesActivationRecord(t *testing.T) {
	store := &memoryStore{}
	svc := New(provider.NewRegistry(providerStub{
		subscribeResp: domain.ProviderSubscribeResult{
			SubscriptionRequestID: "SUBREQ-1",
			ActivationToken:       "provider-token",
			Status:                domain.SubscriptionStatusPendingActivation,
		},
	}), store, config.Config{FrontendBaseURL: "http://localhost:5173"})

	result, err := svc.Subscribe(context.Background(), domain.SubscribeCommand{
		UserID:   "user-123",
		MSISDN:   "6281234567890",
		Provider: domain.ProviderNetplay,
		Plan:     domain.PlanPremium30D,
	})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if result.Record.SubscriptionRequestID != "SUBREQ-1" {
		t.Fatalf("expected SUBREQ-1, got %s", result.Record.SubscriptionRequestID)
	}
	if store.record.ProviderToken != "provider-token" {
		t.Fatalf("expected provider token to be stored")
	}
	if len(store.record.ActivationCode) != 6 {
		t.Fatalf("expected 6-char activation code, got %q", store.record.ActivationCode)
	}
	if store.record.CreatedAt.IsZero() || store.record.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestActivateUpdatesStoredRecord(t *testing.T) {
	activatedAt := time.Date(2026, 5, 17, 16, 29, 33, 0, time.UTC)
	store := &memoryStore{
		record: domain.SubscriptionRecord{
			ActivationCode: "ABC123",
			Provider:       domain.ProviderNetplay,
			ProviderToken:  "provider-token",
		},
	}
	svc := New(provider.NewRegistry(providerStub{
		activateResp: domain.ProviderActivateResult{
			Provider:            domain.ProviderNetplay,
			UserID:              "user-123",
			ActivationStatus:    domain.ActivationStatusSuccess,
			SubscriptionStatus:  domain.SubscriptionStatusActive,
			Plan:                domain.PlanPremium30D,
			ExternalReferenceID: "EXT-1",
			ActivatedAt:         &activatedAt,
			Message:             "Subscription activated successfully",
		},
	}), store, config.Config{FrontendBaseURL: "http://localhost:5173"})

	result, err := svc.Activate(context.Background(), domain.ActivateCommand{ActivationCode: "ABC123"})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}

	if result.Record.SubscriptionStatus != domain.SubscriptionStatusActive {
		t.Fatalf("expected active status, got %s", result.Record.SubscriptionStatus)
	}
	if store.record.ExternalReferenceID != "EXT-1" {
		t.Fatalf("expected external reference to be stored")
	}
}

func TestSubscriptionStatusRefreshesStoredRecord(t *testing.T) {
	activatedAt := time.Date(2026, 5, 17, 16, 29, 33, 0, time.UTC)
	expiresAt := time.Date(2026, 5, 20, 16, 29, 4, 0, time.UTC)
	store := &memoryStore{
		record: domain.SubscriptionRecord{
			ActivationCode: "ABC123",
			Provider:       domain.ProviderNetplay,
			ProviderToken:  "provider-token",
		},
	}
	svc := New(provider.NewRegistry(providerStub{
		statusResp: domain.ProviderStatusResult{
			SubscriptionRequestID: "SUBREQ-1",
			UserID:                "user-123",
			Provider:              domain.ProviderNetplay,
			Plan:                  domain.PlanPremium30D,
			SubscriptionStatus:    domain.SubscriptionStatusActive,
			ActivatedAt:           &activatedAt,
			TokenExpiresAt:        &expiresAt,
			ExternalReferenceID:   "EXT-1",
			Message:               "Subscription is active",
		},
	}), store, config.Config{FrontendBaseURL: "http://localhost:5173"})

	result, err := svc.SubscriptionStatus(context.Background(), domain.StatusCommand{ActivationCode: "ABC123"})
	if err != nil {
		t.Fatalf("SubscriptionStatus() error = %v", err)
	}

	if result.Record.TokenExpiresAt == nil || result.Record.TokenExpiresAt.Format(time.RFC3339) != expiresAt.Format(time.RFC3339) {
		t.Fatalf("expected token expiry to be stored")
	}
}
