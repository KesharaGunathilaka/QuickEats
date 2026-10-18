package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quickeats/pkg/httputil"
	"github.com/quickeats/pkg/middleware"
	"github.com/quickeats/services/payment/internal/app"
	"github.com/quickeats/services/payment/internal/domain"
	"github.com/quickeats/services/payment/internal/port"
)

func NewRouter(paymentUC *app.CreatePaymentUseCase, paymentRepo port.PaymentRepository, walletRepo port.WalletRepository, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "payment"})
	})

	// Internal (service-to-service)
	r.Post("/internal/payments/intent", func(w http.ResponseWriter, r *http.Request) {
		var req app.CreateIntentRequest
		if err := httputil.DecodeJSON(r, &req); err != nil {
			httputil.RespondError(w, err)
			return
		}
		resp, err := paymentUC.Execute(r.Context(), req)
		if err != nil {
			httputil.RespondError(w, err)
			return
		}
		httputil.RespondJSON(w, http.StatusCreated, resp)
	})

	r.Post("/internal/payments/{paymentId}/refund", func(w http.ResponseWriter, r *http.Request) {
		// Refund implementation
		httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "PROCESSING"})
	})

	// Webhook (from payment gateway)
	r.Post("/api/v1/payments/webhook", func(w http.ResponseWriter, r *http.Request) {
		// Parse gateway webhook, verify signature, update payment status
		// This would call paymentUC.completePayment or publishPaymentFailed
		w.WriteHeader(http.StatusOK)
	})

	// Customer routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))

		r.Get("/wallet", func(w http.ResponseWriter, r *http.Request) {
			claims := middleware.GetClaims(r.Context())
			wallet, err := getOrCreateWallet(r, walletRepo, claims.UserID)
			if err != nil {
				httputil.RespondError(w, httputil.ErrNotFound("wallet"))
				return
			}
			httputil.RespondJSON(w, http.StatusOK, wallet)
		})

		r.Post("/wallet/topup", func(w http.ResponseWriter, r *http.Request) {
			claims := middleware.GetClaims(r.Context())
			var req struct{ Amount int64 `json:"amount"` }
			if err := httputil.DecodeJSON(r, &req); err != nil || req.Amount <= 0 {
				httputil.RespondError(w, httputil.ErrValidation("amount must be a positive number of cents", nil))
				return
			}
			if _, err := getOrCreateWallet(r, walletRepo, claims.UserID); err != nil {
				httputil.RespondError(w, err)
				return
			}
			walletRepo.Credit(r.Context(), claims.UserID, req.Amount)
			httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Get("/payments/history", func(w http.ResponseWriter, r *http.Request) {
			// List payments by customer
			httputil.RespondJSON(w, http.StatusOK, []interface{}{})
		})
	})

	return r
}

// getOrCreateWallet lazily creates an empty wallet the first time a customer uses it.
func getOrCreateWallet(r *http.Request, walletRepo port.WalletRepository, userID string) (*domain.Wallet, error) {
	wallet, err := walletRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}
	wallet = domain.NewWallet(userID)
	if err := walletRepo.Create(r.Context(), wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}
