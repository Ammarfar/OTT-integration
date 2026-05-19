package netplay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ott-integration/backend/internal/domain"
	providerdto "ott-integration/backend/internal/dto/provider"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Name() domain.ProviderName {
	return domain.ProviderNetplay
}

func (c *Client) Plans() []domain.Plan {
	return []domain.Plan{domain.PlanPremium30D}
}

func (c *Client) Subscribe(ctx context.Context, cmd domain.SubscribeCommand, idempotencyKey string) (domain.ProviderSubscribeResult, error) {
	payload := providerdto.SubscribeRequest{
		UserID:   cmd.UserID,
		MSISDN:   cmd.MSISDN,
		Provider: string(cmd.Provider),
		Plan:     string(cmd.Plan),
	}

	var response providerdto.SubscribeResponse
	headers := map[string]string{
		"Content-Type":    "application/json",
		"Idempotency-Key": idempotencyKey,
	}

	if err := c.doJSON(ctx, http.MethodPost, "/subscribe", payload, headers, &response); err != nil {
		return domain.ProviderSubscribeResult{}, err
	}

	return domain.ProviderSubscribeResult{
		SubscriptionRequestID: response.SubscriptionRequestID,
		ActivationToken:       response.ActivationToken,
		Status:                domain.SubscriptionStatus(response.Status),
	}, nil
}

func (c *Client) Activate(ctx context.Context, activationToken string) (domain.ProviderActivateResult, error) {
	payload := providerdto.ActivateRequest{ActivationToken: activationToken}
	var response providerdto.ActivateResponse

	if err := c.doJSON(ctx, http.MethodPost, "/activate", payload, map[string]string{"Content-Type": "application/json"}, &response); err != nil {
		return domain.ProviderActivateResult{}, err
	}

	activatedAt, err := parseOptionalTime(response.ActivatedAt)
	if err != nil {
		return domain.ProviderActivateResult{}, fmt.Errorf("parse activatedAt: %w", err)
	}

	return domain.ProviderActivateResult{
		Provider:            domain.ProviderName(response.Provider),
		UserID:              response.UserID,
		ActivationStatus:    domain.ActivationStatus(response.ActivationStatus),
		SubscriptionStatus:  domain.SubscriptionStatus(response.SubscriptionStatus),
		Plan:                domain.Plan(response.Plan),
		ExternalReferenceID: response.ExternalReferenceID,
		ActivatedAt:         activatedAt,
		Message:             response.Message,
	}, nil
}

func (c *Client) GetSubscriptionStatus(ctx context.Context, activationToken string) (domain.ProviderStatusResult, error) {
	query := url.Values{}
	query.Set("activationToken", activationToken)

	var response providerdto.SubscriptionStatusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/subscription-status?"+query.Encode(), nil, nil, &response); err != nil {
		return domain.ProviderStatusResult{}, err
	}

	activatedAt, err := parseOptionalTime(response.ActivatedAt)
	if err != nil {
		return domain.ProviderStatusResult{}, fmt.Errorf("parse activatedAt: %w", err)
	}

	tokenExpiresAt, err := parseOptionalTime(response.TokenExpiresAt)
	if err != nil {
		return domain.ProviderStatusResult{}, fmt.Errorf("parse tokenExpiresAt: %w", err)
	}

	return domain.ProviderStatusResult{
		SubscriptionRequestID: response.SubscriptionRequestID,
		UserID:                response.UserID,
		Provider:              domain.ProviderName(response.Provider),
		Plan:                  domain.Plan(response.Plan),
		SubscriptionStatus:    domain.SubscriptionStatus(response.SubscriptionStatus),
		ActivatedAt:           activatedAt,
		TokenExpiresAt:        tokenExpiresAt,
		ExternalReferenceID:   response.ExternalReferenceID,
		Message:               response.Message,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, requestBody any, headers map[string]string, out any) error {
	var body io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call provider: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read provider response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("provider returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode provider response: %w", err)
	}

	return nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
