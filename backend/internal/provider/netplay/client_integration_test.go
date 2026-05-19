package netplay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ott-integration/backend/internal/domain"
)

func TestClientSubscribeActivateAndStatus(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /subscribe, got %s", r.Method)
		}
		if got := r.Header.Get("Idempotency-Key"); got == "" {
			t.Fatalf("expected idempotency key header")
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"subscriptionRequestId": "SUBREQ-1",
			"activationToken":       "TOKEN-1",
			"status":                "pending_activation",
		})
	})

	mux.HandleFunc("/activate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /activate, got %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"provider":            "NETPLAY",
			"userId":              "user-123",
			"activationStatus":    "success",
			"subscriptionStatus":  "active",
			"plan":                "PREMIUM_30D",
			"externalReferenceId": "EXT-1",
			"activatedAt":         "2026-05-17T16:29:33Z",
			"message":             "Subscription activated successfully",
		})
	})

	mux.HandleFunc("/subscription-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET /subscription-status, got %s", r.Method)
		}
		if got := r.URL.Query().Get("activationToken"); got != "TOKEN-1" {
			t.Fatalf("expected activationToken query, got %s", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"subscriptionRequestId": "SUBREQ-1",
			"userId":                "user-123",
			"provider":              "NETPLAY",
			"plan":                  "PREMIUM_30D",
			"subscriptionStatus":    "active",
			"activatedAt":           "2026-05-17T16:29:33Z",
			"tokenExpiresAt":        "2026-05-20T16:29:04Z",
			"externalReferenceId":   "EXT-1",
			"message":               "Subscription is active",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := New(server.URL, time.Second)

	subscribeResult, err := client.Subscribe(context.Background(), domain.SubscribeCommand{
		UserID:   "user-123",
		MSISDN:   "6281234567890",
		Provider: domain.ProviderNetplay,
		Plan:     domain.PlanPremium30D,
	}, "idempotency-key")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	if subscribeResult.SubscriptionRequestID != "SUBREQ-1" {
		t.Fatalf("expected subscription request id, got %s", subscribeResult.SubscriptionRequestID)
	}

	activateResult, err := client.Activate(context.Background(), "TOKEN-1")
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if activateResult.SubscriptionStatus != domain.SubscriptionStatusActive {
		t.Fatalf("expected active subscription status, got %s", activateResult.SubscriptionStatus)
	}
	if activateResult.ActivatedAt == nil || activateResult.ActivatedAt.Format(time.RFC3339) != "2026-05-17T16:29:33Z" {
		t.Fatalf("expected activatedAt to be parsed")
	}

	statusResult, err := client.GetSubscriptionStatus(context.Background(), "TOKEN-1")
	if err != nil {
		t.Fatalf("GetSubscriptionStatus() error = %v", err)
	}
	if statusResult.TokenExpiresAt == nil || statusResult.TokenExpiresAt.Format(time.RFC3339) != "2026-05-20T16:29:04Z" {
		t.Fatalf("expected tokenExpiresAt to be parsed")
	}
}
