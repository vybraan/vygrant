package storage

import "golang.org/x/oauth2"

type TokenStore interface {
	Set(account string, token *oauth2.Token) error
	Get(account string) (*oauth2.Token, error)
	Delete(account string) error
	ListAccounts() []string
}

type TokenDumper interface {
	Dump() ([]byte, error)
	Restore(data []byte) error
}
