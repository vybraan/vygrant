package daemon

import (
	"context"
	"log"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

func RefreshToken(account string, cfg *config.Config, oldToken *oauth2.Token) (*oauth2.Token, error) {
	acct := cfg.Accounts[account]
	if acct == nil {
		return nil, ErrAccountNotFound
	}

	oauthCfg := config.GetOAuth2Config(acct)

	ts := oauthCfg.TokenSource(context.Background(), oldToken)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func checkExpiringTokens(cfg *config.Config, tokenStore storage.TokenStore) {
	for account := range cfg.Accounts {
		token, err := tokenStore.Get(account)
		if err != nil || token == nil {
			continue
		}

		if token.RefreshToken == "" {
			continue
		}

		if token.Expiry.Before(time.Now().Add(expiryThreshold)) {
			newToken, err := RefreshToken(account, cfg, token)
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
