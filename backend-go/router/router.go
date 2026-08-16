package router

import (
	"net/http"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(healthH *handler.HealthHandler, userH *handler.UserHandler, rideH *handler.RideHandler) http.Handler {
	r := chi.NewRouter()

	// Middlewares padrão
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Healthcheck
	r.Get("/health", healthH.HealthCheck)

	// API v1 Sub-router
	r.Route("/api/v1", func(r chi.Router) {
		// Usuários
		r.Post("/users", userH.CreateUser)

		// Corridas
		r.Post("/rides/estimate", rideH.EstimateRide)
		r.Post("/rides", rideH.CreateRide)
		r.Post("/rides/{id}/accept", rideH.AcceptRide)
		r.Post("/rides/{id}/complete", rideH.CompleteRide)
		r.Get("/rides/{id}", rideH.GetRideDetails)
	})

	return r
}
