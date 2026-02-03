package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	defaultKeyringService = "vygrant"
	accountIndexKey       = "vygrant-account-index"
)

type KeyringStore struct {
	service string
}

type keyringTokenEntry struct {
	RefreshToken string `json:"refresh_token"`
}

func NewKeyringStore(service string) *KeyringStore {
	if service == "" {
		service = defaultKeyringService
	}
	return &KeyringStore{service: service}
}

func KeyringAvailable(service string) bool {
	if service == "" {
		service = defaultKeyringService
	}
	_, err := keyring.Get(service, "vygrant-keyring-check")
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return true
	}
	return false
}

func (k *KeyringStore) Get(account string) (*oauth2.Token, error) {
	secret, err := keyring.Get(k.service, account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	var entry keyringTokenEntry
	if err := json.Unmarshal([]byte(secret), &entry); err == nil {
		if entry.RefreshToken == "" {
			return nil, os.ErrNotExist
		}
		return &oauth2.Token{RefreshToken: entry.RefreshToken}, nil
	}

	if secret == "" {
		return nil, os.ErrNotExist
	}
	return &oauth2.Token{RefreshToken: secret}, nil
}

func (k *KeyringStore) Set(account string, token *oauth2.Token) error {
	if token == nil || token.RefreshToken == "" {
		return nil
	}
	data, err := json.Marshal(keyringTokenEntry{RefreshToken: token.RefreshToken})
	if err != nil {
		return err
	}
	if err := keyring.Set(k.service, account, string(data)); err != nil {
		return err
	}
	return k.addAccountToIndex(account)
}

func (k *KeyringStore) Delete(account string) error {
	deleteErr := keyring.Delete(k.service, account)
	if deleteErr != nil && !errors.Is(deleteErr, keyring.ErrNotFound) {
		return deleteErr
	}
	if err := k.removeAccountFromIndex(account); err != nil {
		return err
	}
	if deleteErr != nil && errors.Is(deleteErr, keyring.ErrNotFound) {
		return os.ErrNotExist
	}
	return nil
}

func (k *KeyringStore) ListAccounts() []string {
	accounts, err := k.readAccountIndex()
	if err != nil {
		return []string{}
	}
	return accounts
}

func (k *KeyringStore) readAccountIndex() ([]string, error) {
	secret, err := keyring.Get(k.service, accountIndexKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return []string{}, nil
		}
		return nil, err
	}
	if secret == "" {
		return []string{}, nil
	}
	var accounts []string
	if err := json.Unmarshal([]byte(secret), &accounts); err != nil {
		return nil, err
	}
	if accounts == nil {
		return []string{}, nil
	}
	return uniqueAccounts(accounts), nil
}

func (k *KeyringStore) writeAccountIndex(accounts []string) error {
	data, err := json.Marshal(uniqueAccounts(accounts))
	if err != nil {
		return err
	}
	return keyring.Set(k.service, accountIndexKey, string(data))
}

func (k *KeyringStore) addAccountToIndex(account string) error {
	if account == "" {
		return nil
	}
	accounts, err := k.readAccountIndex()
	if err != nil {
		return err
	}
	for _, existing := range accounts {
		if existing == account {
			return nil
		}
	}
	accounts = append(accounts, account)
	return k.writeAccountIndex(accounts)
}

func (k *KeyringStore) removeAccountFromIndex(account string) error {
	if account == "" {
		return nil
	}
	accounts, err := k.readAccountIndex()
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return nil
	}
	filtered := make([]string, 0, len(accounts))
	for _, existing := range accounts {
		if existing != account {
			filtered = append(filtered, existing)
		}
	}
	return k.writeAccountIndex(filtered)
}

func uniqueAccounts(accounts []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(accounts))
	for _, account := range accounts {
		if account == "" {
			continue
		}
		if _, ok := seen[account]; ok {
			continue
		}
		seen[account] = struct{}{}
		unique = append(unique, account)
	}
	if unique == nil {
		return []string{}
	}
	return unique
}