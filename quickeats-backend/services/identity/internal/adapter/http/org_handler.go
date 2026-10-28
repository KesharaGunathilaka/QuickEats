package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/services/identity/internal/app"
)

// OrgHandler serves organisation management for admins and the public code lookup.
type OrgHandler struct {
	orgsUC *app.OrganisationsUseCase
}

func NewOrgHandler(orgsUC *app.OrganisationsUseCase) *OrgHandler {
	return &OrgHandler{orgsUC: orgsUC}
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req OrgRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	org, err := h.orgsUC.Create(r.Context(), app.OrgInput{Name: req.Name, Slug: req.Slug, EmailDomains: req.EmailDomains})
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, org)
}

func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req OrgRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	org, err := h.orgsUC.Update(r.Context(), chi.URLParam(r, "id"), app.OrgInput{Name: req.Name, Slug: req.Slug, EmailDomains: req.EmailDomains})
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgsUC.List(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, orgs)
}

// Lookup lets the sign-up screens check a code and show which emails it accepts.
func (h *OrgHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	org, err := h.orgsUC.Lookup(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		respondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, org)
}
