package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

func TestCheckExpiringTokensRefreshesExpiredToken(t *testing.T) {
	tokenEndpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "refresh" {
			t.Fatalf("unexpected refresh request: %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer tokenEndpoint.Close()

	store := storage.NewMemoryStore()
	if err := store.Set("acct", &oauth2.Token{
		AccessToken:  "old-access",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(-time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	checkExpiringTokens(&config.Config{
		Accounts: map[string]*config.Account{
			"acct": {TokenURI: tokenEndpoint.URL},
		},
	}, store, tokenEndpoint.Client())

	token, err := store.Get("acct")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "new-access" {
		t.Fatalf("AccessToken = %q, want new-access", token.AccessToken)
	}
}
