package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vybraan/vygrant/internal/auth"
)

func Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", auth.HandleOAuthCallback)
	r.Get("/auth", auth.StartAuthFlow)

	return r
}
