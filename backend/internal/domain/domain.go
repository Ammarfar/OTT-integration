package domain

import "time"

type ProviderName string

const (
	ProviderNetplay ProviderName = "NETPLAY"
)

type Plan string

const (
	PlanPremium30D Plan = "PREMIUM_30D"
)

type SubscriptionStatus string

const (
	SubscriptionStatusPendingActivation SubscriptionStatus = "pending_activation"
	SubscriptionStatusActive            SubscriptionStatus = "active"
	SubscriptionStatusInactive          SubscriptionStatus = "inactive"
)

type ActivationStatus string

const (
	ActivationStatusPending ActivationStatus = "pending"
	ActivationStatusSuccess ActivationStatus = "success"
	ActivationStatusFailed  ActivationStatus = "failed"
)

type SubscribeCommand struct {
	UserID   string
	MSISDN   string
	Provider ProviderName
	Plan     Plan
}

type SubscribeResult struct {
	Record        SubscriptionRecord
	ActivationURL string
	SMSMessage    string
}

type AppConfig struct {
	FrontendBaseURL string
	HTTPTimeout     time.Duration
}

type ActivateCommand struct {
	ActivationCode string
}

type ActivateResult struct {
	Record SubscriptionRecord
}

type StatusCommand struct {
	ActivationCode string
}

type StatusResult struct {
	Record SubscriptionRecord
}

type ProviderCapability struct {
	Name  ProviderName
	Plans []Plan
}

type SubscriptionRecord struct {
	ActivationCode        string             `json:"activationCode"`
	SubscriptionRequestID string             `json:"subscriptionRequestId"`
	UserID                string             `json:"userId"`
	MSISDN                string             `json:"msisdn"`
	Provider              ProviderName       `json:"provider"`
	Plan                  Plan               `json:"plan"`
	ProviderToken         string             `json:"providerToken"`
	SubscriptionStatus    SubscriptionStatus `json:"subscriptionStatus"`
	ActivationStatus      ActivationStatus   `json:"activationStatus"`
	ExternalReferenceID   string             `json:"externalReferenceId,omitempty"`
	ActivatedAt           *time.Time         `json:"activatedAt,omitempty"`
	TokenExpiresAt        *time.Time         `json:"tokenExpiresAt,omitempty"`
	Message               string             `json:"message"`
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`
}

type ProviderSubscribeResult struct {
	SubscriptionRequestID string
	ActivationToken       string
	Status                SubscriptionStatus
}

type ProviderActivateResult struct {
	Provider            ProviderName
	UserID              string
	ActivationStatus    ActivationStatus
	SubscriptionStatus  SubscriptionStatus
	Plan                Plan
	ExternalReferenceID string
	ActivatedAt         *time.Time
	Message             string
}

type ProviderStatusResult struct {
	SubscriptionRequestID string
	UserID                string
	Provider              ProviderName
	Plan                  Plan
	SubscriptionStatus    SubscriptionStatus
	ActivatedAt           *time.Time
	TokenExpiresAt        *time.Time
	ExternalReferenceID   string
	Message               string
}
