package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/payment/internal/app"
	"github.com/quickeats/services/payment/internal/domain"
)

func NewRouter(paymentUC *app.CreatePaymentUseCase, cardUC *app.CardUseCase, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "payment"})
	})

	// Internal (service-to-service; Kong does not route /internal)
	r.Post("/internal/payments/intent", func(w http.ResponseWriter, r *http.Request) {
		var req app.CreateIntentRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondError(w, err)
			return
		}
		resp, err := paymentUC.Execute(r.Context(), req)
		if err != nil {
			httputil.RespondError(w, paymentError(err))
			return
		}
		httputil.RespondJSON(w, http.StatusCreated, resp)
	})

	// Refunds the order's payment to the customer's card (no-op if it was never charged).
	r.Post("/internal/payments/refund", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OrderID    string `json:"orderId"`
			CustomerID string `json:"customerId"`
		}
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondError(w, err)
			return
		}
		refunded, err := paymentUC.RefundOrder(r.Context(), req.OrderID, req.CustomerID)
		if err != nil {
			httputil.RespondError(w, err)
			return
		}
		httputil.RespondJSON(w, http.StatusOK, map[string]bool{"refunded": refunded})
	})

	// Lets the order service refuse an order up front when the customer can't pay.
	r.Get("/internal/cards/{userId}/check", func(w http.ResponseWriter, r *http.Request) {
		respondPayCheck(w, r, cardUC, chi.URLParam(r, "userId"))
	})

	// Customer routes
	r.Route("/api/v1/cards", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))

		// The customer's saved card, or null. The demo balance is never included.
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			card, err := cardUC.GetCard(r.Context(), middleware.GetClaims(r.Context()).UserID)
			if err != nil {
				httputil.RespondError(w, err)
				return
			}
			httputil.RespondJSON(w, http.StatusOK, card)
		})

		// Adds a card, replacing the current one.
		r.Put("/me", func(w http.ResponseWriter, r *http.Request) {
			var req app.AddCardRequest
			if err := httputil.DecodeJSON(r, &req); err != nil {
				httputil.RespondError(w, err)
				return
			}
			card, err := cardUC.AddCard(r.Context(), middleware.GetClaims(r.Context()).UserID, req)
			var verr *domain.CardValidationError
			if errors.As(err, &verr) {
				httputil.RespondError(w, httputil.ErrValidation(verr.Message, map[string]string{"field": verr.Field}))
				return
			}
			if err != nil {
				httputil.RespondError(w, err)
				return
			}
			httputil.RespondJSON(w, http.StatusOK, card)
		})

		r.Delete("/me", func(w http.ResponseWriter, r *http.Request) {
			if err := cardUC.RemoveCard(r.Context(), middleware.GetClaims(r.Context()).UserID); err != nil {
				httputil.RespondError(w, err)
				return
			}
			httputil.RespondNoContent(w)
		})

		// Whether the saved card can pay ?amount= (cents), without revealing the balance.
		r.Get("/me/check", func(w http.ResponseWriter, r *http.Request) {
			respondPayCheck(w, r, cardUC, middleware.GetClaims(r.Context()).UserID)
		})
	})

	return r
}

func respondPayCheck(w http.ResponseWriter, r *http.Request, cardUC *app.CardUseCase, userID string) {
	amount, err := strconv.ParseInt(r.URL.Query().Get("amount"), 10, 64)
	if err != nil || amount < 0 {
		httputil.RespondError(w, httputil.ErrValidation("amount must be a whole number of cents", nil))
		return
	}
	check, err := cardUC.CanPay(r.Context(), userID, amount)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, check)
}

// paymentError turns a declined charge into a 422 the order service can read the
// reason from (error.details.reason).
func paymentError(err error) error {
	var declined *app.DeclinedError
	if !errors.As(err, &declined) {
		return err
	}
	code, msg := "PAYMENT_DECLINED", "the payment was declined"
	switch declined.Reason {
	case domain.FailureNoCard:
		code, msg = "NO_CARD", "add a card to pay for your order"
	case domain.FailureInsufficientFunds:
		code, msg = "INSUFFICIENT_FUNDS", "your card doesn't have enough balance for this order"
	}
	appErr := httputil.ErrBusinessRule(code, msg)
	appErr.Details = map[string]string{"reason": declined.Reason}
	return appErr
}
