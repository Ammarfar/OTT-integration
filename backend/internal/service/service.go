package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/domain"
	"ott-integration/backend/internal/provider"
	"ott-integration/backend/internal/store"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnsupported        = errors.New("unsupported provider or plan")
	ErrActivationNotFound = errors.New("activation code not found")
)

type Service struct {
	providers *provider.Registry
	store     store.ActivationStore
	config    config.Config
}

func New(providers *provider.Registry, store store.ActivationStore, cfg config.Config) *Service {
	return &Service{
		providers: providers,
		store:     store,
		config:    cfg,
	}
}

func (s *Service) Subscribe(ctx context.Context, cmd domain.SubscribeCommand) (domain.SubscribeResult, error) {
	if err := validateSubscribeCommand(cmd); err != nil {
		return domain.SubscribeResult{}, err
	}

	client, ok := s.providers.Get(cmd.Provider)
	if !ok || !supportsPlan(client.Plans(), cmd.Plan) {
		return domain.SubscribeResult{}, ErrUnsupported
	}

	providerResult, err := client.Subscribe(ctx, cmd, randomKey(16))
	if err != nil {
		return domain.SubscribeResult{}, err
	}

	now := time.Now().UTC()
	record := domain.SubscriptionRecord{
		ActivationCode:        randomKey(6),
		SubscriptionRequestID: providerResult.SubscriptionRequestID,
		UserID:                cmd.UserID,
		MSISDN:                cmd.MSISDN,
		Provider:              cmd.Provider,
		Plan:                  cmd.Plan,
		ProviderToken:         providerResult.ActivationToken,
		SubscriptionStatus:    providerResult.Status,
		ActivationStatus:      domain.ActivationStatusPending,
		Message:               "Activation pending",
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.store.Save(record); err != nil {
		return domain.SubscribeResult{}, fmt.Errorf("save activation record: %w", err)
	}

	link := strings.TrimRight(s.config.FrontendBaseURL, "/") + "/activation/" + record.ActivationCode
	return domain.SubscribeResult{
		Record:        record,
		ActivationURL: link,
		SMSMessage:    "Your NETPLAY activation link: " + link,
	}, nil
}

func (s *Service) Activate(ctx context.Context, cmd domain.ActivateCommand) (domain.ActivateResult, error) {
	if strings.TrimSpace(cmd.ActivationCode) == "" {
		return domain.ActivateResult{}, fmt.Errorf("%w: activationCode is required", ErrInvalidInput)
	}

	record, err := s.store.GetByCode(cmd.ActivationCode)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return domain.ActivateResult{}, ErrActivationNotFound
		}
		return domain.ActivateResult{}, fmt.Errorf("get activation record: %w", err)
	}

	client, ok := s.providers.Get(record.Provider)
	if !ok {
		return domain.ActivateResult{}, ErrUnsupported
	}

	providerResult, err := client.Activate(ctx, record.ProviderToken)
	if err != nil {
		return domain.ActivateResult{}, err
	}

	record.ActivationStatus = providerResult.ActivationStatus
	record.SubscriptionStatus = providerResult.SubscriptionStatus
	record.ExternalReferenceID = providerResult.ExternalReferenceID
	record.ActivatedAt = providerResult.ActivatedAt
	record.Message = providerResult.Message

	if err := s.store.Update(record); err != nil {
		return domain.ActivateResult{}, fmt.Errorf("update activation record: %w", err)
	}

	return domain.ActivateResult{Record: record}, nil
}

func (s *Service) SubscriptionStatus(ctx context.Context, cmd domain.StatusCommand) (domain.StatusResult, error) {
	if strings.TrimSpace(cmd.ActivationCode) == "" {
		return domain.StatusResult{}, fmt.Errorf("%w: activationCode is required", ErrInvalidInput)
	}

	record, err := s.store.GetByCode(cmd.ActivationCode)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return domain.StatusResult{}, ErrActivationNotFound
		}
		return domain.StatusResult{}, fmt.Errorf("get activation record: %w", err)
	}

	client, ok := s.providers.Get(record.Provider)
	if !ok {
		return domain.StatusResult{}, ErrUnsupported
	}

	providerResult, err := client.GetSubscriptionStatus(ctx, record.ProviderToken)
	if err != nil {
		return domain.StatusResult{}, err
	}

	record.SubscriptionStatus = providerResult.SubscriptionStatus
	record.ExternalReferenceID = providerResult.ExternalReferenceID
	record.ActivatedAt = providerResult.ActivatedAt
	record.TokenExpiresAt = providerResult.TokenExpiresAt
	record.Message = providerResult.Message

	if err := s.store.Update(record); err != nil {
		return domain.StatusResult{}, fmt.Errorf("update activation record: %w", err)
	}

	return domain.StatusResult{Record: record}, nil
}

func (s *Service) Providers() []domain.ProviderCapability {
	return s.providers.List()
}

func validateSubscribeCommand(cmd domain.SubscribeCommand) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: userId is required", ErrInvalidInput)
	}
	if strings.TrimSpace(cmd.MSISDN) == "" {
		return fmt.Errorf("%w: msisdn is required", ErrInvalidInput)
	}
	if strings.TrimSpace(string(cmd.Provider)) == "" {
		return fmt.Errorf("%w: provider is required", ErrInvalidInput)
	}
	if strings.TrimSpace(string(cmd.Plan)) == "" {
		return fmt.Errorf("%w: plan is required", ErrInvalidInput)
	}
	if cmd.Provider != domain.ProviderNetplay && cmd.Provider != domain.ProviderNetflix {
		return ErrUnsupported
	}
	if cmd.Plan != domain.PlanPremium30D && cmd.Plan != domain.PlanBasic30D {
		return ErrUnsupported
	}
	return nil
}

const activationAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("X", length)
	}

	var builder strings.Builder
	builder.Grow(length)
	for _, v := range b {
		builder.WriteByte(activationAlphabet[int(v)%len(activationAlphabet)])
	}
	return builder.String()
}

func supportsPlan(plans []domain.Plan, requested domain.Plan) bool {
	for _, plan := range plans {
		if plan == requested {
			return true
		}
	}
	return false
}
