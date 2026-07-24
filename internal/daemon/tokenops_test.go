package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
	"golang.org/x/oauth2"
)

func TestCheckExpiringTokensRefreshesExpiredToken(t *testing.T) {
	tokenEndpoint := newTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		writeTokenResponse(w, "new-access")
	})
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

func TestCheckExpiringTokensRefreshesBeforeNextCheck(t *testing.T) {
	tokenEndpoint := newTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		writeTokenResponse(w, "new-access")
	})
	defer tokenEndpoint.Close()

	store := storage.NewMemoryStore()
	if err := store.Set("acct", &oauth2.Token{
		AccessToken:  "old-access",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(32 * time.Minute),
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

func TestRefreshTokenRetriesRetryableError(t *testing.T) {
	oldDelay := refreshRetryDelay
	refreshRetryDelay = 0
	defer func() { refreshRetryDelay = oldDelay }()

	var calls int32
	tokenEndpoint := newTokenEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			http.Error(w, "try again", http.StatusTooManyRequests)
			return
		}
		writeTokenResponse(w, "new-access")
	})
	defer tokenEndpoint.Close()

	token, err := RefreshToken("acct", &config.Config{
		Accounts: map[string]*config.Account{
			"acct": {TokenURI: tokenEndpoint.URL},
		},
	}, &oauth2.Token{RefreshToken: "refresh"}, tokenEndpoint.Client())
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "new-access" {
		t.Fatalf("AccessToken = %q, want new-access", token.AccessToken)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestRefreshTokenRejectsNilToken(t *testing.T) {
	_, err := RefreshToken("acct", &config.Config{
		Accounts: map[string]*config.Account{
			"acct": {},
		},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func newTokenEndpoint(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "refresh" {
			t.Fatalf("unexpected refresh request: %v", r.Form)
		}
		handler(w, r)
	}))
}

func writeTokenResponse(w http.ResponseWriter, accessToken string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": "refresh",
		"token_type":    "Bearer",
		"expires_in":    3600,
	})
}
