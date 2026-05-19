package providerdto

type SubscribeRequest struct {
	UserID   string `json:"userId"`
	MSISDN   string `json:"msisdn"`
	Provider string `json:"provider"`
	Plan     string `json:"plan"`
}

type SubscribeResponse struct {
	SubscriptionRequestID string `json:"subscriptionRequestId"`
	ActivationToken       string `json:"activationToken"`
	Status                string `json:"status"`
}

type ActivateRequest struct {
	ActivationToken string `json:"activationToken"`
}

type ActivateResponse struct {
	Provider            string `json:"provider"`
	UserID              string `json:"userId"`
	ActivationStatus    string `json:"activationStatus"`
	SubscriptionStatus  string `json:"subscriptionStatus"`
	Plan                string `json:"plan"`
	ExternalReferenceID string `json:"externalReferenceId"`
	ActivatedAt         string `json:"activatedAt"`
	Message             string `json:"message"`
}

type SubscriptionStatusResponse struct {
	SubscriptionRequestID string `json:"subscriptionRequestId"`
	UserID                string `json:"userId"`
	Provider              string `json:"provider"`
	Plan                  string `json:"plan"`
	SubscriptionStatus    string `json:"subscriptionStatus"`
	ActivatedAt           string `json:"activatedAt"`
	TokenExpiresAt        string `json:"tokenExpiresAt"`
	ExternalReferenceID   string `json:"externalReferenceId"`
	Message               string `json:"message"`
}
