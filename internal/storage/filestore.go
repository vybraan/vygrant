package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"
)

type FileStore struct {
	mu     sync.Mutex
	path   string
	tokens map[string]*oauth2.Token
}

func NewFileStore(path string) *FileStore {
	fs := &FileStore{
		path:   path,
		tokens: make(map[string]*oauth2.Token),
	}
	fs.load()
	return fs
}

func (fs *FileStore) Get(account string) (*oauth2.Token, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	token, ok := fs.tokens[account]
	if !ok {
		return nil, os.ErrNotExist
	}
	return token, nil
}

func (fs *FileStore) Set(account string, token *oauth2.Token) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.tokens[account] = token
	return fs.persist()
}

func (fs *FileStore) Delete(account string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.tokens, account)
	return fs.persist()
}

func (fs *FileStore) persist() error {
	tmp := fs.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(fs.path), 0700); err != nil {
		return err
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(fs.tokens); err != nil {
		return err
	}
	return os.Rename(tmp, fs.path)
}

func (fs *FileStore) load() {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		return // empty store
	}
	_ = json.Unmarshal(data, &fs.tokens)
}

func (fs *FileStore) ListAccounts() []string {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	accounts := make([]string, 0, len(fs.tokens))
	for acc := range fs.tokens {
		accounts = append(accounts, acc)
	}
	return accounts
}
