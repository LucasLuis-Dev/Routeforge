package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/client"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/handler"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/repository/postgres"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/router"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/service"
)

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	portStr := getEnvOrDefault("SERVER_PORT", "8080")
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPortStr := getEnvOrDefault("DB_PORT", "5433")
	dbUser := getEnvOrDefault("DB_USER", "routeforge_user")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "routeforge_password")
	dbName := getEnvOrDefault("DB_NAME", "routeforge_db")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")
	mlServiceURL := getEnvOrDefault("ML_SERVICE_URL", "http://localhost:8000")

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		dbPort = 5433
	}

	log.Printf("🔌 Conectando ao PostgreSQL em %s:%d/%s...", dbHost, dbPort, dbName)
	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	})
	if err != nil {
		log.Fatalf("❌ Erro fatal ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()
	log.Println("✅ Conexão com o PostgreSQL estabelecida com sucesso!")

	// Repositórios
	userRepo := postgres.NewUserRepository(db)
	rideRepo := postgres.NewRideRepository(db)

	// Cliente ML (com 2s de timeout)
	mlClient := client.NewMLClient(mlServiceURL, 2*time.Second)

	// Serviços
	rideService := service.NewRideService(userRepo, rideRepo, mlClient)

	// Handlers HTTP
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userRepo)
	rideHandler := handler.NewRideHandler(rideService)

	// Roteador Chi
	r := router.NewRouter(healthHandler, userHandler, rideHandler)

	serverAddr := fmt.Sprintf(":%s", portStr)
	log.Printf("🚀 Servidor Routeforge API iniciado na porta %s (ML URL: %s)", portStr, mlServiceURL)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("❌ Erro no servidor HTTP: %v", err)
	}
}
