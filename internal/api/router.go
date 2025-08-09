package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vybraan/vygrant/internal/auth"
	"github.com/vybraan/vygrant/internal/storage"
)

func Router(tokenStore *storage.TokenStore) http.Handler {
	r := chi.NewRouter()

	r.Get("/", auth.HandleOAuthCallback(*tokenStore))
	r.Get("/auth", auth.StartAuthFlow)

	return r
}
