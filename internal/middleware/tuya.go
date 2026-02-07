package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/avagenc/zee-api/pkg/api"
	"github.com/go-chi/chi/v5"
)

type contextKey int

const (
	tuyaUIDKey contextKey = iota
)

type TuyaUIDResolver interface {
	GetTuyaUIDByOwnerID(ctx context.Context, ownerID string) (string, error)
}

type DeviceLister interface {
	GetUserDeviceIDs(tuyaUID string) ([]string, error)
}

type Tuya struct {
	resolver TuyaUIDResolver
	lister   DeviceLister
}

func NewTuya(resolver TuyaUIDResolver, lister DeviceLister) *Tuya {
	return &Tuya{resolver: resolver, lister: lister}
}

func (t *Tuya) RequireAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ownerID, err := api.GetUserIDFromContext(r.Context())
		if err != nil {
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "Missing user identity", nil))
			return
		}

		tuyaUID, err := t.resolver.GetTuyaUIDByOwnerID(r.Context(), ownerID)
		if err != nil {
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "No Tuya App Account is linked to the user", nil))
			return
		}

		ctx := newContextWithTuyaUID(r.Context(), tuyaUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (t *Tuya) RequireDeviceOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tuyaUID, err := GetTuyaUIDFromContext(r.Context())
		if err != nil {
			api.Respond(w, http.StatusUnauthorized, api.NewErrorResponse("UNAUTHORIZED", "Tuya identity not found", nil))
			return
		}

		deviceID := chi.URLParam(r, "deviceId")
		if deviceID == "" {
			api.Respond(w, http.StatusBadRequest, api.NewErrorResponse("BAD_REQUEST", "Device ID is required", nil))
			return
		}

		deviceIDs, err := t.lister.GetUserDeviceIDs(tuyaUID)
		if err != nil {
			api.Respond(w, http.StatusInternalServerError, api.NewErrorResponse("INTERNAL_ERROR", "Failed to verify device ownership", nil))
			return
		}

		if !contains(deviceIDs, deviceID) {
			api.Respond(w, http.StatusForbidden, api.NewErrorResponse("FORBIDDEN", "Device does not belong to user", nil))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newContextWithTuyaUID(ctx context.Context, tuyaUID string) context.Context {
	return context.WithValue(ctx, tuyaUIDKey, tuyaUID)
}

func GetTuyaUIDFromContext(ctx context.Context) (string, error) {
	val, ok := ctx.Value(tuyaUIDKey).(string)
	if !ok || val == "" {
		return "", errors.New("tuya UID not found in context")
	}
	return val, nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
