package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quickeats/pkg/middleware"
)

func NewRouter(handler *CatalogHandler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "catalog"})
	})

	// Customer routes (authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Get("/shops", handler.ListShops)
		r.Get("/shops/{id}", handler.GetShop)
		r.Get("/shops/{id}/menu", handler.ListMenu)

		// Shop staff routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("shop_staff", "shop_owner", "admin"))
			r.Post("/shops/{id}/menu-items", handler.CreateMenuItem)
			r.Patch("/menu-items/{itemId}/stock", handler.UpdateStock)
		})
	})

	// Internal routes (service-to-service, no JWT)
	r.Route("/internal", func(r chi.Router) {
		r.Post("/stock/reserve", handler.ReserveStock)
		r.Post("/stock/release", handler.ReleaseStock)
	})

	return r
}
