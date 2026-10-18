package http

import (
	"net/http"

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

func (h *CatalogHandler) GetShop(w http.ResponseWriter, r *http.Request) {
	shop, err := h.browseUC.GetShop(r.Context(), chi.URLParam(r, "id"))
	if err != nil || shop == nil {
		httputil.RespondError(w, httputil.ErrNotFound("shop"))
		return
	}
	httputil.RespondJSON(w, http.StatusOK, shop)
}

func (h *CatalogHandler) ListMenu(w http.ResponseWriter, r *http.Request) {
	items, err := h.menuUC.ListItems(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, items)
}

func (h *CatalogHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var req CreateMenuItemRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	item, err := h.menuUC.CreateItem(r.Context(), chi.URLParam(r, "id"), req.Name, req.Description, req.Price, req.Category, req.Stock)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, item)
}

func (h *CatalogHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var req UpdateStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}
	if err := h.menuUC.UpdateStock(r.Context(), chi.URLParam(r, "itemId"), req.Stock); err != nil {
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
	shop := domain.NewShop(claims.OrganisationID, req.Name, req.Description, claims.UserID)
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
	if req.OrganisationID == "" || req.OwnerID == "" || req.Name == "" {
		httputil.RespondError(w, httputil.ErrValidation("organisationId, ownerId and name are required", nil))
		return
	}
	shop := domain.NewShop(req.OrganisationID, req.Name, req.Description, req.OwnerID)
	shop.IsOpen = true // the shop app has no open/close toggle yet, so new shops start open
	if err := h.shopRepo.Create(r.Context(), shop); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusCreated, shop)
}

func (h *CatalogHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	if err := h.menuUC.DeleteItem(r.Context(), chi.URLParam(r, "itemId")); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondNoContent(w)
}
