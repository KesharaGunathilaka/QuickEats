package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/catalog/internal/app"
	"github.com/quickeats/services/catalog/internal/domain"
	"github.com/quickeats/services/catalog/internal/port"
)

type CatalogHandler struct {
	browseUC *app.BrowseShopsUseCase
	menuUC   *app.ManageMenuUseCase
	stockUC  *app.ReserveStockUseCase
	shopRepo port.ShopRepository
}

func NewCatalogHandler(browseUC *app.BrowseShopsUseCase, menuUC *app.ManageMenuUseCase, stockUC *app.ReserveStockUseCase, shopRepo port.ShopRepository) *CatalogHandler {
	return &CatalogHandler{browseUC: browseUC, menuUC: menuUC, stockUC: stockUC, shopRepo: shopRepo}
}

func (h *CatalogHandler) ListShops(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		httputil.RespondError(w, httputil.ErrUnauthenticated("no claims"))
		return
	}
	shops, err := h.browseUC.ListByOrg(r.Context(), claims.OrganisationID)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, shops)
}

// visibleShop loads a shop the caller may see: shops serving their organisation, their
// own shop, or any shop for an admin. Other shops look like they don't exist.
func (h *CatalogHandler) visibleShop(w http.ResponseWriter, r *http.Request, shopID string) *domain.Shop {
	claims := middleware.GetClaims(r.Context())
	shop, err := h.browseUC.GetShop(r.Context(), shopID)
	if err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return nil
	}
	if claims.Role == "admin" || claims.ShopID == shop.ID || shop.ServesOrganisation(claims.OrganisationID) {
		return shop
	}
	httputil.RespondError(w, httputil.ErrNotFound("shop"))
	return nil
}

func (h *CatalogHandler) GetShop(w http.ResponseWriter, r *http.Request) {
	shop := h.visibleShop(w, r, chi.URLParam(r, "id"))
	if shop == nil {
		return
	}
	httputil.RespondJSON(w, http.StatusOK, shop)
}

// UpdateShop lets a merchant edit their own shop's details and open/closed state.
func (h *CatalogHandler) UpdateShop(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	shopID := chi.URLParam(r, "id")
	if claims.Role != "admin" && claims.ShopID != shopID {
		httputil.RespondError(w, httputil.ErrForbidden("you can only edit your own shop"))
		return
	}
	var req UpdateShopRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	shop, err := h.shopRepo.GetByID(r.Context(), shopID)
	if err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 255 {
			httputil.RespondError(w, httputil.ErrValidation("shop name must be 1-255 characters", nil))
			return
		}
		shop.Name = name
	}
	if req.Description != nil {
		shop.Description = strings.TrimSpace(*req.Description)
	}
	if req.IsOpen != nil {
		shop.IsOpen = *req.IsOpen
	}
	if err := h.shopRepo.Update(r.Context(), shop); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, shop)
}

func (h *CatalogHandler) ListMenu(w http.ResponseWriter, r *http.Request) {
	if h.visibleShop(w, r, chi.URLParam(r, "id")) == nil {
		return
	}
	items, err := h.menuUC.ListItems(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, items)
}

func (h *CatalogHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	shopID := chi.URLParam(r, "id")
	if claims.Role != "admin" && claims.ShopID != shopID {
		httputil.RespondError(w, httputil.ErrForbidden("you can only manage your own shop's menu"))
		return
	}
	var req CreateMenuItemRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	item, err := h.menuUC.CreateItem(r.Context(), shopID, req.Name, req.Description, req.Price, req.Category, req.Stock)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, item)
}

func (h *CatalogHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	itemID := chi.URLParam(r, "itemId")
	item, err := h.menuUC.GetItem(r.Context(), itemID)
	if err != nil || item == nil || (claims.Role != "admin" && claims.ShopID != item.ShopID) {
		httputil.RespondError(w, httputil.ErrNotFound("menu item"))
		return
	}
	var req UpdateStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	if err := h.menuUC.UpdateStock(r.Context(), itemID, req.Stock); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]int{"stock": req.Stock})
}

func (h *CatalogHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	var req ReserveStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	items := make([]port.StockReservationItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = port.StockReservationItem{MenuItemID: it.MenuItemID, Quantity: it.Quantity}
	}
	result, failed, err := h.stockUC.Reserve(r.Context(), req.OrderID, items)
	if err != nil {
		if failed != nil {
			httputil.RespondError(w, httputil.ErrBusinessRule("STOCK_INSUFFICIENT", "not enough stock"))
			return
		}
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, result)
}

// CommitStock is called by the order service when an order is confirmed.
func (h *CatalogHandler) CommitStock(w http.ResponseWriter, r *http.Request) {
	var req CommitStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	committed, err := h.stockUC.Commit(r.Context(), req.OrderID)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]int{"committed": committed})
}

func (h *CatalogHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	var req ReleaseStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	released, err := h.stockUC.Release(r.Context(), req.OrderID, req.ReservationIDs)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]int{"released": released})
}

func (h *CatalogHandler) CreateShop(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req CreateShopRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	if req.Name == "" {
		httputil.RespondError(w, httputil.ErrValidation("name is required", nil))
		return
	}
	// Simplified: create directly via repo
	// In production, this would go through a use case
	shop := domain.NewShop([]string{claims.OrganisationID}, req.Name, req.Description, claims.UserID)
	if err := h.shopRepo.Create(r.Context(), shop); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, shop)
}

// CreateShopInternal is the service-to-service endpoint identity calls during shop sign-up.
func (h *CatalogHandler) CreateShopInternal(w http.ResponseWriter, r *http.Request) {
	var req InternalCreateShopRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	orgIDs := req.OrganisationIDs
	if len(orgIDs) == 0 && req.OrganisationID != "" {
		orgIDs = []string{req.OrganisationID}
	}
	if len(orgIDs) == 0 || req.OwnerID == "" || req.Name == "" {
		httputil.RespondError(w, httputil.ErrValidation("organisationIds, ownerId and name are required", nil))
		return
	}
	shop := domain.NewShop(orgIDs, req.Name, req.Description, req.OwnerID)
	shop.IsOpen = true // new shops start open; the merchant can close it in settings
	if err := h.shopRepo.Create(r.Context(), shop); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, shop)
}

// ShopOrganisationsInternal returns the organisations a shop serves (used by identity).
func (h *CatalogHandler) ShopOrganisationsInternal(w http.ResponseWriter, r *http.Request) {
	shop, err := h.shopRepo.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{"organisationIds": shop.OrganisationIDs})
}

// SetShopOrganisationsInternal replaces the organisations a shop serves (used by identity).
func (h *CatalogHandler) SetShopOrganisationsInternal(w http.ResponseWriter, r *http.Request) {
	var req InternalShopOrganisationsRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	if len(req.OrganisationIDs) == 0 {
		httputil.RespondError(w, httputil.ErrValidation("organisationIds is required", nil))
		return
	}
	shopID := chi.URLParam(r, "id")
	if shop, err := h.shopRepo.GetByID(r.Context(), shopID); err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return
	}
	if err := h.shopRepo.SetOrganisations(r.Context(), shopID, req.OrganisationIDs); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{"organisationIds": req.OrganisationIDs})
}

// ShopVisibleInternal tells the order service whether a customer's organisation may order from a shop.
func (h *CatalogHandler) ShopVisibleInternal(w http.ResponseWriter, r *http.Request) {
	shop, err := h.shopRepo.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]bool{
		"visible": shop.ServesOrganisation(r.URL.Query().Get("organisationId")),
		"isOpen":  shop.IsOpen,
	})
}

func (h *CatalogHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	itemID := chi.URLParam(r, "itemId")
	item, err := h.menuUC.GetItem(r.Context(), itemID)
	if err != nil || item == nil || (claims.Role != "admin" && claims.ShopID != item.ShopID) {
		httputil.RespondError(w, httputil.ErrNotFound("menu item"))
		return
	}
	if err := h.menuUC.DeleteItem(r.Context(), itemID); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondNoContent(w)
}
