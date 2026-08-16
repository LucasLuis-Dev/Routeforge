package router

import (
	"net/http"
	"time"

	_ "github.com/LucasLuis-Dev/Routeforge/backend-go/docs"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/handler"
	customMw "github.com/LucasLuis-Dev/Routeforge/backend-go/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/time/rate"
)

func NewRouter(
	healthH *handler.HealthHandler,
	userH *handler.UserHandler,
	rideH *handler.RideHandler,
	authH *handler.AuthHandler,
	driverH *handler.DriverHandler,
) http.Handler {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Rate Limiter para estimativas: 10 requisições por minuto por IP (burst = 10)
	estimateLimiter := customMw.NewIPRateLimiter(rate.Every(6*time.Second), 10)

	// Swagger UI Documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Healthcheck
	r.Get("/health", healthH.HealthCheck)

	// API v1 Sub-router
	r.Route("/api/v1", func(r chi.Router) {
		// Rotas Públicas
		r.Post("/users", userH.CreateUser)
		r.Post("/auth/login", authH.Login)

		// Rota de Estimativa (Pública com Rate Limiting por IP)
		r.With(estimateLimiter.Limit).Post("/rides/estimate", rideH.EstimateRide)

		// Rotas Autenticadas via Token JWT
		r.Group(func(r chi.Router) {
			r.Use(customMw.AuthMiddleware)

			// Atualização de Localização GPS em Tempo Real (Restrito a Motoristas)
			r.With(customMw.RequireRole(domain.UserTypeDriver)).Post("/drivers/location", driverH.UpdateLocation)

			// Consulta de Motoristas Próximos (Redis GEOSEARCH)
			r.Get("/drivers/nearby", driverH.FindNearby)

			// Solicitar corrida (Restrito a Passageiros)
			r.With(customMw.RequireRole(domain.UserTypePassenger)).Post("/rides", rideH.CreateRide)

			// Aceitar corrida (Restrito a Motoristas)
			r.With(customMw.RequireRole(domain.UserTypeDriver)).Post("/rides/{id}/accept", rideH.AcceptRide)

			// Finalizar e Consultar corrida (Autenticados)
			r.Post("/rides/{id}/complete", rideH.CompleteRide)
			r.Get("/rides/{id}", rideH.GetRideDetails)
		})
	})

	return r
}
