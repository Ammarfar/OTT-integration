package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ott-integration/backend/internal/domain"
)

var ErrRecordNotFound = errors.New("activation record not found")

type FileStore struct {
	path    string
	mu      sync.Mutex
	records map[string]domain.SubscriptionRecord
}

func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{
		path:    path,
		records: make(map[string]domain.SubscriptionRecord),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *FileStore) Save(record domain.SubscriptionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[record.ActivationCode] = record
	return s.persistLocked()
}

func (s *FileStore) GetByCode(code string) (domain.SubscriptionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[code]
	if !ok {
		return domain.SubscriptionRecord{}, ErrRecordNotFound
	}

	return record, nil
}

func (s *FileStore) Update(record domain.SubscriptionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[record.ActivationCode]; !ok {
		return ErrRecordNotFound
	}

	record.UpdatedAt = time.Now().UTC()
	s.records[record.ActivationCode] = record
	return s.persistLocked()
}

func (s *FileStore) load() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read store: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var records []domain.SubscriptionRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode store: %w", err)
	}

	for _, record := range records {
		s.records[record.ActivationCode] = record
	}

	return nil
}

func (s *FileStore) persistLocked() error {
	records := make([]domain.SubscriptionRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal store: %w", err)
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp store: %w", err)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace store: %w", err)
	}

	return nil
}
