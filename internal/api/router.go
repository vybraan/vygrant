package api

import (
	"net/http"

	"github.com/vybraan/vygrant/internal/auth"
	"github.com/vybraan/vygrant/internal/storage"
)

// Router creates an HTTP router configured with routes for the OAuth callback (GET "/")
// and authentication initiation (GET "/auth"). The OAuth callback handler is provided the
// given tokenStore and httpClient. It returns the configured http.Handler.
func Router(tokenStore storage.TokenStore, httpClient *http.Client) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", auth.HandleOAuthCallback(tokenStore, httpClient))
	mux.HandleFunc("GET /auth", auth.StartAuthFlow)

	return mux
}
