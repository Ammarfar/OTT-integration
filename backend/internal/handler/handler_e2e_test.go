package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ott-integration/backend/internal/config"
	"ott-integration/backend/internal/domain"
	"ott-integration/backend/internal/provider"
	"ott-integration/backend/internal/service"
	"ott-integration/backend/internal/store"
)

type e2eProviderStub struct {
	subscribeResp domain.ProviderSubscribeResult
	activateResp  domain.ProviderActivateResult
	statusResp    domain.ProviderStatusResult
}

func (p e2eProviderStub) Name() domain.ProviderName {
	return domain.ProviderNetplay
}

func (p e2eProviderStub) Plans() []domain.Plan {
	return []domain.Plan{domain.PlanPremium30D}
}

func (p e2eProviderStub) Subscribe(_ context.Context, _ domain.SubscribeCommand, _ string) (domain.ProviderSubscribeResult, error) {
	return p.subscribeResp, nil
}

func (p e2eProviderStub) Activate(_ context.Context, _ string) (domain.ProviderActivateResult, error) {
	return p.activateResp, nil
}

func (p e2eProviderStub) GetSubscriptionStatus(_ context.Context, _ string) (domain.ProviderStatusResult, error) {
	return p.statusResp, nil
}

func TestBackendE2EFlow(t *testing.T) {
	activatedAt := time.Date(2026, 5, 17, 16, 29, 33, 0, time.UTC)
	expiresAt := time.Date(2026, 5, 20, 16, 29, 4, 0, time.UTC)

	dir := t.TempDir()
	fileStore, err := store.NewFileStore(filepath.Join(dir, "activations.json"))
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	svc := service.New(
		provider.NewRegistry(e2eProviderStub{
			subscribeResp: domain.ProviderSubscribeResult{
				SubscriptionRequestID: "SUBREQ-1",
				ActivationToken:       "TOKEN-1",
				Status:                domain.SubscriptionStatusPendingActivation,
			},
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
		}),
		fileStore,
		config.Config{
			FrontendBaseURL: "http://frontend.local",
			HTTPTimeout:     time.Second,
		},
	)

	router := NewRouter(svc, time.Second)

	subscribeReq := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(`{"userId":"user-123","msisdn":"6281234567890","provider":"NETPLAY","plan":"PREMIUM_30D"}`))
	subscribeReq.Header.Set("Content-Type", "application/json")
	subscribeRec := httptest.NewRecorder()
	router.ServeHTTP(subscribeRec, subscribeReq)

	if subscribeRec.Code != http.StatusOK {
		t.Fatalf("subscribe status = %d", subscribeRec.Code)
	}

	var subscribeResp struct {
		SubscriptionRequestID string `json:"subscriptionRequestId"`
		ActivationCode        string `json:"activationCode"`
		ActivationLink        string `json:"activationLink"`
		SMSMessage            string `json:"smsMessage"`
		Status                string `json:"status"`
	}
	if err := json.Unmarshal(subscribeRec.Body.Bytes(), &subscribeResp); err != nil {
		t.Fatalf("unmarshal subscribe response: %v", err)
	}
	if subscribeResp.SubscriptionRequestID != "SUBREQ-1" {
		t.Fatalf("expected subscription request id SUBREQ-1, got %s", subscribeResp.SubscriptionRequestID)
	}
	if subscribeResp.ActivationCode == "" {
		t.Fatalf("expected activation code")
	}
	if subscribeResp.ActivationLink != "http://frontend.local/activation/"+subscribeResp.ActivationCode {
		t.Fatalf("unexpected activation link: %s", subscribeResp.ActivationLink)
	}

	activateReq := httptest.NewRequest(http.MethodPost, "/api/activate", strings.NewReader(`{"activationCode":"`+subscribeResp.ActivationCode+`"}`))
	activateReq.Header.Set("Content-Type", "application/json")
	activateRec := httptest.NewRecorder()
	router.ServeHTTP(activateRec, activateReq)

	if activateRec.Code != http.StatusOK {
		t.Fatalf("activate status = %d", activateRec.Code)
	}

	var activateResp struct {
		Provider            string `json:"provider"`
		UserID              string `json:"userId"`
		ActivationStatus    string `json:"activationStatus"`
		SubscriptionStatus  string `json:"subscriptionStatus"`
		Plan                string `json:"plan"`
		ExternalReferenceID string `json:"externalReferenceId"`
		ActivatedAt         string `json:"activatedAt"`
		Message             string `json:"message"`
	}
	if err := json.Unmarshal(activateRec.Body.Bytes(), &activateResp); err != nil {
		t.Fatalf("unmarshal activate response: %v", err)
	}
	if activateResp.SubscriptionStatus != "active" {
		t.Fatalf("expected active subscription status, got %s", activateResp.SubscriptionStatus)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/subscription-status?activationCode="+subscribeResp.ActivationCode, nil)
	statusRec := httptest.NewRecorder()
	router.ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status code = %d", statusRec.Code)
	}

	var statusResp struct {
		SubscriptionRequestID string `json:"subscriptionRequestId"`
		SubscriptionStatus    string `json:"subscriptionStatus"`
		TokenExpiresAt        string `json:"tokenExpiresAt"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("unmarshal status response: %v", err)
	}
	if statusResp.SubscriptionRequestID != "SUBREQ-1" {
		t.Fatalf("expected subscription request id SUBREQ-1, got %s", statusResp.SubscriptionRequestID)
	}
	if statusResp.SubscriptionStatus != "active" {
		t.Fatalf("expected active subscription status, got %s", statusResp.SubscriptionStatus)
	}

	providersReq := httptest.NewRequest(http.MethodGet, "/api/providers", nil)
	providersRec := httptest.NewRecorder()
	router.ServeHTTP(providersRec, providersReq)
	if providersRec.Code != http.StatusOK {
		t.Fatalf("providers status = %d", providersRec.Code)
	}
}
