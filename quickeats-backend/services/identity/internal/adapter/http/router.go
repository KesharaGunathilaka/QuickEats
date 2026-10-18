package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quickeats/pkg/middleware"
)

func NewRouter(authHandler *AuthHandler, orgHandler *OrgHandler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "identity"})
	})

	// Public auth routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/register-shop", authHandler.RegisterShop)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Admin routes
	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Use(middleware.RequireRole("admin"))
		r.Post("/organisations", orgHandler.Create)
		r.Get("/organisations", orgHandler.List)
		r.Get("/organisations/{id}", orgHandler.GetByID)
	})

	// Internal service-to-service validation
	r.Route("/internal/auth", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Get("/validate", func(w http.ResponseWriter, r *http.Request) {
			claims := middleware.GetClaims(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"userId":         claims.UserID,
					"organisationId": claims.OrganisationID,
					"role":           claims.Role,
					"shopId":         claims.ShopID,
				},
			})
		})
	})

	return r
}
