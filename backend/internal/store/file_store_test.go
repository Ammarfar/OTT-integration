package store

import (
	"path/filepath"
	"testing"

	"ott-integration/backend/internal/domain"
)

func TestFileStoreSaveLoadUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activations.json")

	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	record := domain.SubscriptionRecord{
		ActivationCode:        "ABC123",
		SubscriptionRequestID: "SUBREQ-1",
		ProviderToken:         "provider-token",
	}

	if err := store.Save(record); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.GetByCode("ABC123")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if got.SubscriptionRequestID != "SUBREQ-1" {
		t.Fatalf("expected persisted record")
	}

	got.ProviderToken = "updated-token"
	if err := store.Update(got); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := store.GetByCode("ABC123")
	if err != nil {
		t.Fatalf("GetByCode() after update error = %v", err)
	}
	if updated.ProviderToken != "updated-token" {
		t.Fatalf("expected updated provider token")
	}
}
