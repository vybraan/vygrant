package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"golang.org/x/oauth2"
)

const keyringSecretPrefix = "v1:"

func EncodeTokenSecret(token *oauth2.Token) (string, error) {
	if token == nil || token.RefreshToken == "" {
		return "", errors.New("empty refresh token")
	}
	data, err := json.Marshal(keyringTokenEntry{RefreshToken: token.RefreshToken})
	if err != nil {
		return "", err
	}
	return keyringSecretPrefix + base64.StdEncoding.EncodeToString(data), nil
}

func DecodeTokenSecret(secret string) (*oauth2.Token, error) {
	if strings.HasPrefix(secret, keyringSecretPrefix) {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, keyringSecretPrefix))
		if err != nil {
			return nil, err
		}
		var entry keyringTokenEntry
		if err := json.Unmarshal(decoded, &entry); err != nil {
			return nil, err
		}
		if entry.RefreshToken == "" {
			return nil, errors.New("empty refresh token")
		}
		return &oauth2.Token{RefreshToken: entry.RefreshToken}, nil
	}

	trimmed := strings.TrimSpace(secret)
	if strings.HasPrefix(trimmed, "{") {
		var entry keyringTokenEntry
		if err := json.Unmarshal([]byte(secret), &entry); err != nil {
			return nil, err
		}
		if entry.RefreshToken == "" {
			return nil, errors.New("empty refresh token")
		}
		return &oauth2.Token{RefreshToken: entry.RefreshToken}, nil
	}

	if secret == "" {
		return nil, errors.New("empty secret")
	}
	return &oauth2.Token{RefreshToken: secret}, nil
}
