package netflix

import (
	"context"
	"fmt"
	"time"

	"ott-integration/backend/internal/domain"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) Name() domain.ProviderName {
	return domain.ProviderNetflix
}

func (c *Client) Plans() []domain.Plan {
	return []domain.Plan{domain.PlanPremium30D}
}

func (c *Client) Subscribe(_ context.Context, cmd domain.SubscribeCommand, _ string) (domain.ProviderSubscribeResult, error) {
	return domain.ProviderSubscribeResult{
		SubscriptionRequestID: "NFLX-" + randomSuffix(cmd.UserID),
		ActivationToken:       "NFLX-TOKEN-" + randomSuffix(cmd.MSISDN),
		Status:                domain.SubscriptionStatusPendingActivation,
	}, nil
}

func (c *Client) Activate(_ context.Context, activationToken string) (domain.ProviderActivateResult, error) {
	now := time.Now().UTC()
	return domain.ProviderActivateResult{
		Provider:            domain.ProviderNetflix,
		UserID:              "demo-user",
		ActivationStatus:    domain.ActivationStatusSuccess,
		SubscriptionStatus:  domain.SubscriptionStatusActive,
		Plan:                domain.PlanPremium30D,
		ExternalReferenceID: "NFLX-EXT-" + randomSuffix(activationToken),
		ActivatedAt:         &now,
		Message:             "Netflix subscription activated successfully",
	}, nil
}

func (c *Client) GetSubscriptionStatus(_ context.Context, activationToken string) (domain.ProviderStatusResult, error) {
	now := time.Now().UTC()
	expires := now.Add(72 * time.Hour)
	return domain.ProviderStatusResult{
		SubscriptionRequestID: "NFLX-" + randomSuffix(activationToken),
		UserID:                "demo-user",
		Provider:              domain.ProviderNetflix,
		Plan:                  domain.PlanPremium30D,
		SubscriptionStatus:    domain.SubscriptionStatusActive,
		ActivatedAt:           &now,
		TokenExpiresAt:        &expires,
		ExternalReferenceID:   "NFLX-EXT-" + randomSuffix(activationToken),
		Message:               "Netflix subscription is active",
	}, nil
}

func randomSuffix(seed string) string {
	sum := 0
	for _, r := range seed {
		sum += int(r)
	}
	return fmt.Sprintf("%06d", sum%1000000)
}
