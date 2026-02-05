package device

import (
	"encoding/json"
	"net/http"

	"github.com/avagenc/zee-api/pkg/api"
)

type Service interface {
	ListByHome(homeID string) ([]Device, error)
	SendCommands(deviceID string, commands []DataPoint) (json.RawMessage, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListByHome(w http.ResponseWriter, r *http.Request) {
	homeID := r.PathValue("homeId")
	if homeID == "" {
		api.Respond(w, http.StatusBadRequest, api.NewErrorResponse("INVALID_REQUEST", "Missing homeId", nil))
		return
	}

	devices, err := h.svc.ListByHome(homeID)
	if err != nil {
		api.Respond(w, http.StatusBadGateway, api.NewErrorResponse("UPSTREAM_ERROR", err.Error(), nil))
		return
	}

	api.Respond(w, http.StatusOK, api.NewSuccessResponse("Devices retrieved successfully", devices, nil))
}

func (h *Handler) SendCommands(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID string      `json:"device_id"`
		Commands []DataPoint `json:"commands"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Respond(w, http.StatusBadRequest, api.NewErrorResponse("INVALID_REQUEST", "Invalid request body", nil))
		return
	}

	if req.DeviceID == "" {
		api.Respond(w, http.StatusBadRequest, api.NewErrorResponse("INVALID_REQUEST", "Missing device_id", nil))
		return
	}

	if len(req.Commands) == 0 {
		api.Respond(w, http.StatusBadRequest, api.NewErrorResponse("INVALID_REQUEST", "Commands cannot be empty", nil))
		return
	}

	result, err := h.svc.SendCommands(req.DeviceID, req.Commands)
	if err != nil {
		api.Respond(w, http.StatusBadGateway, api.NewErrorResponse("UPSTREAM_ERROR", err.Error(), nil))
		return
	}

	api.Respond(w, http.StatusOK, api.NewSuccessResponse("Commands sent successfully", result, nil))
}
