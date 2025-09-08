package storage

import (
	"errors"
	"sync"

	"golang.org/x/oauth2"
)

type MemoryStore struct {
	tokens map[string]*oauth2.Token
	mu     sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tokens: make(map[string]*oauth2.Token),
	}
}

func (m *MemoryStore) Set(account string, token *oauth2.Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[account] = token

	return nil
}

func (m *MemoryStore) Get(account string) (*oauth2.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, exists := m.tokens[account]
	if !exists {
		return nil, errors.New("token not found")
	}
	return token, nil
}

func (m *MemoryStore) Delete(account string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.tokens[account]; !exists {
		return errors.New("token not found")
	}
	delete(m.tokens, account)
	return nil
}

func (m *MemoryStore) ListAccounts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	accounts := make([]string, 0, len(m.tokens))
	for acc := range m.tokens {
		accounts = append(accounts, acc)
	}
	return accounts
}
