package handlers

import (
	"net/http"

	response "github.com/olad5/file-fort/pkg/utils"
)

func (h HealthHandler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.db.Ping(); err != nil {
		response.ErrorResponse(w, "postgres is down", http.StatusNotFound)
		return
	}

	if err := h.cache.Ping(ctx); err != nil {
		response.ErrorResponse(w, "redis cache is down", http.StatusOK)
		return
	}

	response.ErrorResponse(w, "service is live", http.StatusOK)
	return
}
