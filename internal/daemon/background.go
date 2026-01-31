package daemon

import (
	"log"
	"net/http"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
)

const (
	checkInterval   = 30 * time.Minute
	expiryThreshold = 10 * time.Minute
)

// StartBackgroundTasks starts a background loop that periodically checks for expiring tokens using the provided configuration, token store, and HTTP client.
// It schedules checks at checkInterval and stops gracefully when stopCh is signaled.
func StartBackgroundTasks(cfg *config.Config, tokenStore storage.TokenStore, httpClient *http.Client, stopCh <-chan struct{}) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkExpiringTokens(cfg, tokenStore, httpClient)
		case <-stopCh:
			log.Println("Stopping background tasks...")
			return
		}
	}
}