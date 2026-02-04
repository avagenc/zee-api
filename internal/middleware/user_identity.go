package middleware

import (
	"net/http"

	"github.com/avagenc/zee-api/pkg/api"
)

type UserIdentity struct{}

func NewUserIdentity() *UserIdentity {
	return &UserIdentity{}
}

func (u *UserIdentity) ToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("x-user-id")
		if userID == "" {
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "Missing user identity", nil))
			return
		}

		ctx, err := api.NewContextWithUserID(r.Context(), userID)
		if err != nil {
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "Invalid user identity", nil))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
