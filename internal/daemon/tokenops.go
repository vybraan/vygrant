package daemon

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

const refreshRetries = 1

var refreshRetryDelay = time.Second

// RefreshToken obtains a new OAuth2 token for the named account using the provided existing token.
// If httpClient is non-nil it is attached to the refresh request context and used for HTTP calls.
// It returns ErrAccountNotFound if the account is not present in cfg.Accounts, or any error produced by the token source when fetching the new token.
func RefreshToken(account string, cfg *config.Config, oldToken *oauth2.Token, httpClient *http.Client) (*oauth2.Token, error) {
	acct := cfg.Accounts[account]
	if acct == nil {
		return nil, ErrAccountNotFound
	}
	if oldToken == nil {
		return nil, errors.New("oauth2: token is nil")
	}

	oauthCfg := config.GetOAuth2Config(acct)
	ctx := context.Background()
	if httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}
	token := *oldToken
	token.Expiry = time.Now().Add(-time.Second)
	ts := oauthCfg.TokenSource(ctx, &token)
	var err error
	for attempt := 0; ; attempt++ {
		var newToken *oauth2.Token
		newToken, err = ts.Token()
		if err == nil {
			return newToken, nil
		}
		if attempt == refreshRetries || !retryableRefreshError(err) {
			return nil, err
		}
		time.Sleep(refreshRetryDelay)
	}
}

func retryableRefreshError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) && retrieveErr.Response != nil {
		status := retrieveErr.Response.StatusCode
		return status == http.StatusTooManyRequests || status >= 500
	}
	return false
}

// checkExpiringTokens iterates configured accounts and refreshes tokens whose expiry is within expiryThreshold.
// It skips accounts with no stored token or without a refresh token. For tokens needing refresh it calls
// RefreshToken (using the provided httpClient when non-nil), updates tokenStore on success, and logs and
// notifies on refresh failures.
func checkExpiringTokens(cfg *config.Config, tokenStore storage.TokenStore, httpClient *http.Client) {
	now := time.Now()
	for account := range cfg.Accounts {
		token, err := tokenStore.Get(account)
		if err != nil || token == nil {
			continue
		}

		if token.RefreshToken == "" {
			continue
		}

		if token.Expiry.Before(now.Add(expiryThreshold)) {
			newToken, err := RefreshToken(account, cfg, token, httpClient)
			if err != nil {
				log.Printf("failed to auto-refresh token for %s: %v", account, err)
				Notify("vygrant - auto refresh", "Token for "+account+" could not be refreshed.")
				continue
			}

			if err := tokenStore.Set(account, newToken); err != nil {
				log.Printf("failed to save refreshed token for %s: %v", account, err)
			}

			log.Printf("Token for %s refreshed. New expiry: %s", account, newToken.Expiry)
		}

	}
}
