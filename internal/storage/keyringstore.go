package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const defaultKeyringService = "vygrant"

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
	return keyring.Set(k.service, account, string(data))
}

func (k *KeyringStore) Delete(account string) error {
	if err := keyring.Delete(k.service, account); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return os.ErrNotExist
		}
		return err
	}
	return nil
}

func (k *KeyringStore) ListAccounts() []string {
	return nil
}