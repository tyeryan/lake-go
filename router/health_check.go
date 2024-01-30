package router

import (
	"github.com/go-chi/render"
	logutil "github.com/tyeryan/l-protocol/log"
	"net/http"
)

func (h *LakeHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	log := logutil.GetLogger("HealthCheck")
	ctx := r.Context()
	log.Infow(ctx, "Health check called")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"health": true})
}
