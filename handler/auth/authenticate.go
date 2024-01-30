package auth

import (
	"encoding/json"
	"github.com/go-chi/render"
	"github.com/tyeryan/l-protocol/go/lauth"
	logutil "github.com/tyeryan/l-protocol/log"
	"io"
	"net/http"
)

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	log := logutil.GetLogger("Authenticate")
	var err error
	ctx := r.Context()

	defer func() {
		log.Infow(ctx, "Authenticate done")
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var reqBody AuthenticationReqBody
	if err := json.Unmarshal(body, &reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// call auth client here
	authReq := &lauth.AuthReq{
		Username: reqBody.Username,
		Password: reqBody.Password,
	}

	authRsp, err := h.client.Authenticate(ctx, authReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, authRsp)
}

type AuthenticationReqBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
