package daemon

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

func RefreshToken(account string, cfg *config.Config, oldToken *oauth2.Token, httpClient *http.Client) (*oauth2.Token, error) {
	acct := cfg.Accounts[account]
	if acct == nil {
		return nil, ErrAccountNotFound
	}

	oauthCfg := config.GetOAuth2Config(acct)
	ctx := context.Background()
	if httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}
	ts := oauthCfg.TokenSource(ctx, oldToken)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func checkExpiringTokens(cfg *config.Config, tokenStore storage.TokenStore, httpClient *http.Client) {
	for account := range cfg.Accounts {
		token, err := tokenStore.Get(account)
		if err != nil || token == nil {
			continue
		}

		if token.RefreshToken == "" {
			continue
		}

		if token.Expiry.Before(time.Now().Add(expiryThreshold)) {
			newToken, err := RefreshToken(account, cfg, token, httpClient)
			if err != nil {
				log.Printf("failed to auto-refresh token for %s: %v", account, err)
				Notify("vygrant - auto refresh", "Token for "+account+" could not be refreshed.")
				continue
			}

			tokenStore.Set(account, newToken)

			log.Printf("Token for %s refreshed. New expiry: %s", account, newToken.Expiry)
		} else if token.Expiry.Before(time.Now()) {
			log.Printf("Token for %s has expired. Please refresh manually.", account)
			Notify("vygrant - token expired", "Token for "+account+" has expired and must be refreshed manually.")
		}

	}
}
