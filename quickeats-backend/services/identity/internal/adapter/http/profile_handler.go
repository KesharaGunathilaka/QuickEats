package http

import (
	"net/http"

	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/identity/internal/app"
)

// ProfileHandler serves the signed-in user's own account (the apps' settings screens).
type ProfileHandler struct {
	profileUC *app.ProfileUseCase
}

func NewProfileHandler(profileUC *app.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profileUC: profileUC}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.profileUC.Get(r.Context(), middleware.GetClaims(r.Context()).UserID)
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, p)
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	p, err := h.profileUC.UpdateDisplayName(r.Context(), middleware.GetClaims(r.Context()).UserID, req.DisplayName)
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, p)
}

func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	if err := h.profileUC.ChangePassword(r.Context(), middleware.GetClaims(r.Context()).UserID, req.CurrentPassword, req.NewPassword); err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondNoContent(w)
}

func (h *ProfileHandler) SetShopOrganisations(w http.ResponseWriter, r *http.Request) {
	var req ShopOrganisationsRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	p, err := h.profileUC.SetShopOrganisations(r.Context(), middleware.GetClaims(r.Context()).UserID, req.OrganisationSlugs)
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, p)
}
