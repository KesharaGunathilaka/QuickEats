package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
)

type OrgHandler struct {
	orgRepo port.OrganisationRepository
}

func NewOrgHandler(orgRepo port.OrganisationRepository) *OrgHandler {
	return &OrgHandler{orgRepo: orgRepo}
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrgRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	if req.Name == "" || req.Slug == "" || req.VerifyMethod == "" {
		httputil.RespondError(w, httputil.ErrValidation("name, slug, and verifyMethod are required", nil))
		return
	}

	org := domain.NewOrganisation(req.Name, req.Slug, domain.VerifyMethod(req.VerifyMethod), req.VerifyConfig)
	if err := h.orgRepo.Create(r.Context(), org); err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, org)
}

func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgRepo.List(r.Context())
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, orgs)
}

func (h *OrgHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	org, err := h.orgRepo.GetByID(r.Context(), id)
	if err != nil || org == nil {
		httputil.RespondError(w, httputil.ErrNotFound("organisation"))
		return
	}
	httputil.RespondJSON(w, http.StatusOK, org)
}
