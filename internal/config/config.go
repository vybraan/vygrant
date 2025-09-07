package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"golang.org/x/oauth2"
)

type Account struct {
	AuthURI       string            `toml:"auth_uri"`
	TokenURI      string            `toml:"token_uri"`
	ClientID      string            `toml:"client_id"`
	ClientSecret  string            `toml:"client_secret"`
	RedirectURI   string            `toml:"redirect_uri"`
	Scopes        []string          `toml:"scopes"`
	AuthURIFields map[string]string `toml:"auth_uri_fields"`
}

type Config struct {
	HTTPSListen   string              `toml:"https_listen"`
	HTTPListen    string              `toml:"http_listen"`
	PersistTokens bool                `toml:"persist_tokens"`
	Accounts      map[string]*Account `toml:"account"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s — please run `vygrant init` or create one manually", path)
		}
	}
	return &cfg, nil
}

func LoadConfigFromEnv() (*Config, error) {
	confPath := os.Getenv("VYGRANT_CONFIG")
	if confPath == "" {
		return nil, fmt.Errorf("could not load env config")
	}
	return LoadConfig(confPath)
}

func GetOAuth2Config(acct *Account) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     acct.ClientID,
		ClientSecret: acct.ClientSecret,
		RedirectURL:  acct.RedirectURI,
		Scopes:       acct.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  acct.AuthURI,
			TokenURL: acct.TokenURI,
		},
	}
}
