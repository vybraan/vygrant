package daemon

import (
	"errors"
	"log"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

func StartBackgroundTasks(cfg *config.Config, tokenStore storage.TokenStore) {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				checkExpiringTokens(cfg, tokenStore)
			}
		}
	}()
}

func checkExpiringTokens(cfg *config.Config, tokenStore storage.TokenStore) {
	for account := range cfg.Accounts {
		// token, err := storage.LoadToken(account)
		token, err := tokenStore.Get(account)

		if err != nil {
			continue
		}

		if token.Expiry.Before(time.Now().Add(10 * time.Minute)) {
			// Alert user if token is expiring soon and cannot be auto-refreshed
			// Notify("Token Expiring", "Token for "+account+" is expiring soon and must be manually refreshed.")

			newToken, err := RefreshToken(account, cfg)
			if err != nil {
				log.Printf("Failed to auto-refresh token for %s: %v", account, err)
				Notify("Auto-refresh failed", "Token for "+account+" could not be refreshed.")
				continue
			}
			log.Printf("Token for %s refreshed: expires at %s", account, newToken.Expiry)
		}
	}
}

func RefreshToken(account string, cfg *config.Config) (oauth2.Token, error) {
	return oauth2.Token{}, errors.New("not implemented")
}
