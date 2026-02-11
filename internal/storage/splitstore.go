package storage

import (
	"encoding/json"
	"errors"
	"os"

	"golang.org/x/oauth2"
)

type SplitStore struct {
	access  *MemoryStore
	refresh TokenStore
}

func NewSplitStore(refresh TokenStore) *SplitStore {
	return &SplitStore{
		access:  NewMemoryStore(),
		refresh: refresh,
	}
}

func (s *SplitStore) Get(account string) (*oauth2.Token, error) {
	var token *oauth2.Token

	accessToken, errAccess := s.access.Get(account)
	if errAccess != nil && !isNotExist(errAccess) {
		return nil, errAccess
	}
	if errAccess == nil && accessToken != nil {
		copyToken := *accessToken
		token = &copyToken
	}

	refreshToken, errRefresh := s.refresh.Get(account)
	if errRefresh != nil && !isNotExist(errRefresh) {
		return nil, errRefresh
	}
	if errRefresh == nil && refreshToken != nil && refreshToken.RefreshToken != "" {
		if token == nil {
			token = &oauth2.Token{}
		}
		token.RefreshToken = refreshToken.RefreshToken
	}

	if token == nil {
		return nil, os.ErrNotExist
	}

	return token, nil
}

func (s *SplitStore) Set(account string, token *oauth2.Token) error {
	if token == nil {
		return os.ErrInvalid
	}

	accessCopy := *token
	accessCopy.RefreshToken = ""
	if err := s.access.Set(account, &accessCopy); err != nil {
		return err
	}

	if token.RefreshToken != "" {
		refreshCopy := &oauth2.Token{RefreshToken: token.RefreshToken}
		if err := s.refresh.Set(account, refreshCopy); err != nil {
			return err
		}
	}

	return nil
}

func (s *SplitStore) Delete(account string) error {
	errAccess := s.access.Delete(account)
	errRefresh := s.refresh.Delete(account)

	if errAccess == nil && errRefresh == nil {
		return nil
	}

	if errAccess != nil && !isNotExist(errAccess) {
		return errAccess
	}
	if errRefresh != nil && !isNotExist(errRefresh) {
		return errRefresh
	}

	if errAccess != nil && errRefresh != nil {
		return os.ErrNotExist
	}

	return nil
}

func (s *SplitStore) ListAccounts() []string {
	seen := map[string]struct{}{}
	for _, account := range s.access.ListAccounts() {
		seen[account] = struct{}{}
	}
	for _, account := range s.refresh.ListAccounts() {
		seen[account] = struct{}{}
	}
	accounts := make([]string, 0, len(seen))
	for account := range seen {
		accounts = append(accounts, account)
	}
	return accounts
}

func (s *SplitStore) RefreshStore() TokenStore {
	return s.refresh
}

func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

func (s *SplitStore) Dump() ([]byte, error) {
	type dumpEntry struct {
		Access  *oauth2.Token `json:"access,omitempty"`
		Refresh *oauth2.Token `json:"refresh,omitempty"`
	}
	dump := map[string]dumpEntry{}
	accounts := s.ListAccounts()
	for _, account := range accounts {
		entry := dumpEntry{}
		if token, err := s.access.Get(account); err == nil {
			copyToken := *token
			entry.Access = &copyToken
		}
		if token, err := s.refresh.Get(account); err == nil {
			copyToken := *token
			entry.Refresh = &copyToken
		}
		if entry.Access != nil || entry.Refresh != nil {
			dump[account] = entry
		}
	}
	return json.Marshal(dump)
}

func (s *SplitStore) Restore(data []byte) error {
	type dumpEntry struct {
		Access  *oauth2.Token `json:"access,omitempty"`
		Refresh *oauth2.Token `json:"refresh,omitempty"`
	}
	if len(data) == 0 {
		return nil
	}
	var dump map[string]dumpEntry
	if err := json.Unmarshal(data, &dump); err != nil {
		return err
	}
	for account, entry := range dump {
		if entry.Access != nil {
			if err := s.access.Set(account, entry.Access); err != nil {
				return err
			}
		}
		if entry.Refresh != nil {
			if err := s.refresh.Set(account, entry.Refresh); err != nil {
				return err
			}
		}
	}
	return nil
}
