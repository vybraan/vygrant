package storage

import (
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

func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}