package storage

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

const defaultPassPrefix = "vygrant"

type PassStore struct {
	prefix string
}

func NewPassStore(prefix string) *PassStore {
	if prefix == "" {
		prefix = defaultPassPrefix
	}
	return &PassStore{prefix: prefix}
}

func PassAvailable() bool {
	_, err := exec.LookPath("pass")
	return err == nil
}

func (p *PassStore) Get(account string) (*oauth2.Token, error) {
	entry := p.entryPath(account)
	cmd := exec.Command("pass", "show", entry)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if isPassNotFound(output) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	secret := strings.TrimSpace(string(output))
	token, err := DecodeTokenSecret(secret)
	if err != nil {
		if err.Error() == "empty secret" || err.Error() == "empty refresh token" {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return token, nil
}

func (p *PassStore) Set(account string, token *oauth2.Token) error {
	if token == nil || token.RefreshToken == "" {
		return nil
	}
	entry := p.entryPath(account)
	secret, err := EncodeTokenSecret(token)
	if err != nil {
		return err
	}
	cmd := exec.Command("pass", "insert", "-m", "-f", entry)
	cmd.Stdin = bytes.NewBufferString(secret + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *PassStore) Delete(account string) error {
	entry := p.entryPath(account)
	cmd := exec.Command("pass", "rm", "-f", entry)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if isPassNotFound(output) {
			return os.ErrNotExist
		}
		return err
	}
	return nil
}

func (p *PassStore) ListAccounts() []string {
	cmd := exec.Command("pass", "ls", p.prefix)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if isPassNotFound(output) {
			return []string{}
		}
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	accounts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Password Store") {
			continue
		}
		line = strings.TrimLeft(line, "├└─│ ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, p.prefix) {
			line = strings.TrimPrefix(line, p.prefix)
			line = strings.TrimPrefix(line, "/")
		}
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "/") {
			continue
		}
		if account, err := url.PathUnescape(filepath.Base(line)); err == nil {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func (p *PassStore) entryPath(account string) string {
	escaped := url.PathEscape(account)
	return filepath.Join(p.prefix, escaped)
}

func isPassNotFound(output []byte) bool {
	text := strings.ToLower(string(output))
	return strings.Contains(text, "not in the password store") ||
		strings.Contains(text, "is not in the password store") ||
		strings.Contains(text, "error:") && strings.Contains(text, "not found")
}

func (p *PassStore) Dump() ([]byte, error) {
	accounts := p.ListAccounts()
	dump := make(map[string]*oauth2.Token, len(accounts))
	for _, account := range accounts {
		token, err := p.Get(account)
		if err != nil {
			continue
		}
		if token != nil {
			copyToken := *token
			dump[account] = &copyToken
		}
	}
	return json.Marshal(dump)
}

func (p *PassStore) Restore(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var dump map[string]*oauth2.Token
	if err := json.Unmarshal(data, &dump); err != nil {
		return err
	}
	for account, token := range dump {
		if token == nil {
			continue
		}
		if err := p.Set(account, token); err != nil {
			return err
		}
	}
	return nil
}
