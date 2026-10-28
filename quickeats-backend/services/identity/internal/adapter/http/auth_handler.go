package http

import (
	"net/http"

	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/services/identity/internal/app"
	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
)

type AuthHandler struct {
	registerUC     *app.RegisterUseCase
	registerShopUC *app.RegisterShopUseCase
	loginUC        *app.LoginUseCase
	tokenIssuer    port.TokenIssuer
	userRepo       port.UserRepository
}

func NewAuthHandler(
	registerUC *app.RegisterUseCase,
	registerShopUC *app.RegisterShopUseCase,
	loginUC *app.LoginUseCase,
	tokenIssuer port.TokenIssuer,
	userRepo port.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		registerUC:     registerUC,
		registerShopUC: registerShopUC,
		loginUC:        loginUC,
		tokenIssuer: tokenIssuer,
		userRepo:    userRepo,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" || req.OrganisationSlug == "" {
		httputil.RespondError(w, httputil.ErrValidation("email, password, displayName, and organisationSlug are required", nil))
		return
	}

	user, err := h.registerUC.Execute(r.Context(), app.RegisterRequest{
		OrganisationSlug: req.OrganisationSlug,
		Email:            req.Email,
		Password:         req.Password,
		DisplayName:      req.DisplayName,
		InviteCode:       req.InviteCode,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, user)
}

// RegisterShop signs up a shop owner together with their shop.
func (h *AuthHandler) RegisterShop(w http.ResponseWriter, r *http.Request) {
	var req RegisterShopRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	slugs := req.OrganisationSlugs
	if len(slugs) == 0 && req.OrganisationSlug != "" {
		slugs = []string{req.OrganisationSlug}
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" || len(slugs) == 0 || req.ShopName == "" {
		httputil.RespondError(w, httputil.ErrValidation("email, password, displayName, organisationSlugs, and shopName are required", nil))
		return
	}

	user, err := h.registerShopUC.Execute(r.Context(), app.RegisterShopRequest{
		Email:             req.Email,
		Password:          req.Password,
		DisplayName:       req.DisplayName,
		OrganisationSlugs: slugs,
		ShopName:          req.ShopName,
		ShopDescription:   req.ShopDescription,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	tokens, err := h.loginUC.Execute(r.Context(), app.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			httputil.RespondError(w, httputil.ErrUnauthenticated(err.Error()))
		} else {
			httputil.RespondError(w, err)
		}
		return
	}

	httputil.RespondJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	userID, err := h.tokenIssuer.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		httputil.RespondError(w, httputil.ErrUnauthenticated("invalid refresh token"))
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		httputil.RespondError(w, httputil.ErrUnauthenticated("user not found"))
		return
	}

	tokens, err := h.tokenIssuer.Issue(user)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, tokens)
}
