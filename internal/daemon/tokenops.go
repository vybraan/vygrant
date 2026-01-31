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

// RefreshToken obtains a new OAuth2 token for the named account using the provided existing token.
// If httpClient is non-nil it is attached to the refresh request context and used for HTTP calls.
// It returns ErrAccountNotFound if the account is not present in cfg.Accounts, or any error produced by the token source when fetching the new token.
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

// checkExpiringTokens iterates configured accounts and refreshes tokens whose expiry is within expiryThreshold.
// It skips accounts with no stored token or without a refresh token. For tokens needing refresh it calls
// RefreshToken (using the provided httpClient when non-nil), updates tokenStore on success, and logs and
// notifies on refresh failures. If a token is already expired it logs and sends an expiration notification.
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