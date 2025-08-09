package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const tokenDir = ".vybr/vygrant"

// tokenFilePath returns the full path to the token file for a given account.
func tokenFilePath(account string) (string, error) {
	userConfigDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userConfigDir, tokenDir, account+".json"), nil
}

func SaveToken(account string, token *oauth2.Token) error {
	path, err := tokenFilePath(account)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0700)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func LoadToken(account string) (*oauth2.Token, error) {
	path, err := tokenFilePath(account)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var token oauth2.Token
	err = json.NewDecoder(f).Decode(&token)
	return &token, err
}

func DeleteToken(account string) error {
	path, err := tokenFilePath(account)
	if err != nil {
		return err
	}
	// Check if the file exists before attempting to delete
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("token for account '%s' not found", account)
	}

	return os.Remove(path)
}
