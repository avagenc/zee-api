package system

import (
	"net/http"

	"github.com/avagenc/zee-api/internal/config"
	"github.com/avagenc/zee-api/pkg/api"
)

type Handler struct {
	cfg *config.App
}

func NewHandler(cfg *config.App) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Service     string `json:"service"`
		Status      string `json:"status"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	}{
		Service:     h.cfg.Name,
		Status:      "UP",
		Environment: h.cfg.Env,
		Version:     h.cfg.Version,
	}

	api.Respond(w, http.StatusOK, api.NewSuccessResponse("Service is healthy", data, nil))
}
