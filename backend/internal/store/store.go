package store

import (
	"ott-integration/backend/internal/domain"
)

type ActivationStore interface {
	Save(record domain.SubscriptionRecord) error
	GetByCode(code string) (domain.SubscriptionRecord, error)
	Update(record domain.SubscriptionRecord) error
}
