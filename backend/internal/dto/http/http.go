package httpdto

type SubscribeRequest struct {
	UserID   string `json:"userId"`
	MSISDN   string `json:"msisdn"`
	Provider string `json:"provider"`
	Plan     string `json:"plan"`
}

type SubscribeResponse struct {
	SubscriptionRequestID string `json:"subscriptionRequestId"`
	ActivationCode        string `json:"activationCode"`
	ActivationLink        string `json:"activationLink"`
	SMSMessage            string `json:"smsMessage"`
	Status                string `json:"status"`
}

type ActivateRequest struct {
	ActivationCode string `json:"activationCode"`
}

type ActivateResponse struct {
	Provider            string `json:"provider"`
	UserID              string `json:"userId"`
	ActivationStatus    string `json:"activationStatus"`
	SubscriptionStatus  string `json:"subscriptionStatus"`
	Plan                string `json:"plan"`
	ExternalReferenceID string `json:"externalReferenceId,omitempty"`
	ActivatedAt         string `json:"activatedAt,omitempty"`
	Message             string `json:"message"`
}

type SubscriptionStatusResponse struct {
	SubscriptionRequestID string `json:"subscriptionRequestId"`
	UserID                string `json:"userId"`
	Provider              string `json:"provider"`
	Plan                  string `json:"plan"`
	SubscriptionStatus    string `json:"subscriptionStatus"`
	ActivatedAt           string `json:"activatedAt,omitempty"`
	TokenExpiresAt        string `json:"tokenExpiresAt,omitempty"`
	ExternalReferenceID   string `json:"externalReferenceId,omitempty"`
	Message               string `json:"message"`
}

type ProviderInfo struct {
	Name  string   `json:"name"`
	Plans []string `json:"plans"`
}
