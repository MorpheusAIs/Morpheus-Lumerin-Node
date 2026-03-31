package mobile

import (
	"fmt"
	"sync"
)

// MemoryKeyValueStorage implements interfaces.KeyValueStorage in-memory.
// Suitable for mobile where OS keychain may not be available or desired.
type MemoryKeyValueStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryKeyValueStorage() *MemoryKeyValueStorage {
	return &MemoryKeyValueStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryKeyValueStorage) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return v, nil
}

func (m *MemoryKeyValueStorage) Insert(key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		return fmt.Errorf("key already exists: %s", key)
	}
	m.data[key] = value
	return nil
}

func (m *MemoryKeyValueStorage) Upsert(key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MemoryKeyValueStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}
	delete(m.data, key)
	return nil
}

func (m *MemoryKeyValueStorage) DeleteIfExists(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}
