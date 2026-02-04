package middleware

import (
	"log"
	"net/http"

	"github.com/avagenc/zee-api/internal/config"
	"github.com/avagenc/zee-api/pkg/api"
)

type APIKey struct {
	apiKey string
}

func NewAPIKey(cfg *config.Security) *APIKey {
	return &APIKey{apiKey: cfg.APIKey}
}

func (a *APIKey) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientAPIKey := r.Header.Get("x-avagenc-api-key")
		if clientAPIKey != a.apiKey {
			log.Printf("Authentication failed: Invalid Avagenc API Key. Request from %s", r.RemoteAddr)
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "Unauthorized access", nil))
			return
		}
		next.ServeHTTP(w, r)
	})
}
