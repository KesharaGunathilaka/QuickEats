package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/order/internal/app"
	"github.com/quickeats/services/order/internal/domain"
	"github.com/quickeats/services/order/internal/port"
)

type OrderHandler struct {
	placeOrderUC  *app.PlaceOrderUseCase
	orderStatusUC *app.OrderStatusUseCase
	orderRepo     port.OrderRepository
}

func NewOrderHandler(placeOrderUC *app.PlaceOrderUseCase, orderStatusUC *app.OrderStatusUseCase, orderRepo port.OrderRepository) *OrderHandler {
	return &OrderHandler{placeOrderUC: placeOrderUC, orderStatusUC: orderStatusUC, orderRepo: orderRepo}
}

type PlaceOrderHTTPRequest struct {
	ShopID         string       `json:"shopId"`
	IdempotencyKey string       `json:"idempotencyKey"`
	PaymentMethod  string       `json:"paymentMethod"`
	Items          []OrderItemDTO `json:"items"`
}

type OrderItemDTO struct {
	MenuItemID string `json:"menuItemId"`
	ItemName   string `json:"itemName"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unitPrice"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type RejectRequest struct {
	Reason string `json:"reason"`
}

type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req PlaceOrderHTTPRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	if req.IdempotencyKey == "" || req.ShopID == "" || len(req.Items) == 0 {
		httputil.RespondError(w, httputil.ErrValidation("idempotencyKey, shopId, and items are required", nil))
		return
	}

	var items []domain.OrderItem
	for _, dto := range req.Items {
		items = append(items, domain.OrderItem{
			MenuItemID: dto.MenuItemID, ItemName: dto.ItemName,
			Quantity: dto.Quantity, UnitPrice: dto.UnitPrice,
		})
	}

	order, err := h.placeOrderUC.Execute(r.Context(), app.PlaceOrderRequest{
		CustomerID:     claims.UserID,
		ShopID:         req.ShopID,
		IdempotencyKey: req.IdempotencyKey,
		PaymentMethod:  req.PaymentMethod,
		Items:          items,
	})
	if err != nil {
		if err == domain.ErrDuplicateOrder {
			httputil.RespondError(w, httputil.ErrConflict("order with this idempotency key already exists"))
		} else {
			httputil.RespondError(w, err)
		}
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderRepo.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil || order == nil {
		httputil.RespondError(w, httputil.ErrNotFound("order"))
		return
	}
	httputil.RespondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) ListByCustomer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	params := httputil.ParsePagination(r)
	orders, err := h.orderRepo.ListByCustomer(r.Context(), claims.UserID, params.Limit, params.Cursor)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) ListByShop(w http.ResponseWriter, r *http.Request) {
	shopID := chi.URLParam(r, "shopId")
	params := httputil.ParsePagination(r)
	orders, err := h.orderRepo.ListByShop(r.Context(), shopID, params.Limit, params.Cursor)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateStatusRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	order, err := h.orderRepo.GetByID(r.Context(), id)
	if err != nil || order == nil {
		httputil.RespondError(w, httputil.ErrNotFound("order"))
		return
	}

	oldStatus := order.Status
	if err := order.TransitionTo(domain.OrderStatus(req.Status)); err != nil {
		httputil.RespondError(w, httputil.ErrBusinessRule("INVALID_STATUS_TRANSITION", err.Error()))
		return
	}

	if err := h.orderStatusUC.UpdateStatus(r.Context(), order, oldStatus); err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order, err := h.orderRepo.GetByID(r.Context(), id)
	if err != nil || order == nil {
		httputil.RespondError(w, httputil.ErrNotFound("order"))
		return
	}

	// Check cancellation window
	if order.CancellationDeadline != nil && time.Now().After(*order.CancellationDeadline) {
		httputil.RespondError(w, httputil.ErrBusinessRule("CANCELLATION_EXPIRED", "cancellation window has expired"))
		return
	}

	if err := order.TransitionTo(domain.StatusCancelled); err != nil {
		httputil.RespondError(w, httputil.ErrBusinessRule("INVALID_STATUS_TRANSITION", err.Error()))
		return
	}

	// In production: also trigger refund + stock release via events
	if err := h.orderStatusUC.Cancel(r.Context(), order, "cancelled_by_customer"); err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) RejectOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req RejectRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, err)
		return
	}

	order, err := h.orderRepo.GetByID(r.Context(), id)
	if err != nil || order == nil {
		httputil.RespondError(w, httputil.ErrNotFound("order"))
		return
	}

	if err := order.TransitionTo(domain.StatusRejected); err != nil {
		httputil.RespondError(w, httputil.ErrBusinessRule("INVALID_STATUS_TRANSITION", err.Error()))
		return
	}

	if err := h.orderStatusUC.Reject(r.Context(), order, req.Reason); err != nil {
		httputil.RespondError(w, err)
		return
	}
	order.RejectReason = req.Reason
	httputil.RespondJSON(w, http.StatusOK, order)
}
