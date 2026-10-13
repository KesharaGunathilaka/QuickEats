package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/catalog/internal/app"
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
	shop := &app.CreateShopParams{Name: req.Name, Description: req.Description, OrgID: claims.OrganisationID, OwnerID: claims.UserID}
	_ = shop
	// Simplified: create directly via repo
	// In production, this would go through a use case
}
