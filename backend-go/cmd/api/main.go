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
	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/messaging"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/repository/postgres"
	redisRepo "github.com/LucasLuis-Dev/Routeforge/backend-go/repository/redis"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/router"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/service"
	ws "github.com/LucasLuis-Dev/Routeforge/backend-go/websocket"
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
	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("REDIS_PORT", "6379")
	rabbitmqURL := getEnvOrDefault("RABBITMQ_URL", "amqp://routeforge_user:routeforge_password@localhost:5672/")

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		dbPort = 5433
	}

	log.Printf("Conectando ao PostgreSQL em %s:%d/%s...", dbHost, dbPort, dbName)
	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	})
	if err != nil {
		log.Fatalf("Erro fatal ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()
	log.Println("Conexão com o PostgreSQL estabelecida com sucesso!")

	log.Printf("Conectando ao Redis 7 em %s:%s...", redisHost, redisPort)
	rdb, err := redisRepo.NewRedisClient(redisHost, redisPort)
	if err != nil {
		log.Printf("Aviso: Não foi possível conectar ao Redis (%v). Executando sem cache/GEO.", err)
	} else {
		log.Println("Conexão com o Redis 7 estabelecida com sucesso!")
		defer rdb.Close()
	}

	log.Printf("Conectando ao RabbitMQ em %s...", rabbitmqURL)
	eventPublisher, err := messaging.NewRabbitMQPublisher(rabbitmqURL)
	if err != nil {
		log.Printf("Aviso: Não foi possível conectar ao RabbitMQ (%v). Eventos de mensageria desativados.", err)
	} else {
		defer eventPublisher.Close()
	}

	// Inicializa e executa em background o Hub de WebSockets (Driver Tracking Streaming)
	wsHub := ws.NewHub()
	go wsHub.Run()
	log.Println("WebSocket Hub para Streaming de Localização em Tempo Real iniciado!")

	// Repositórios
	userRepo := postgres.NewUserRepository(db)
	rideRepo := postgres.NewRideRepository(db)

	var geoRepo redisRepo.GeoRepository
	var estimateCache redisRepo.EstimateCache

	if rdb != nil {
		geoRepo = redisRepo.NewGeoRepository(rdb)
		estimateCache = redisRepo.NewEstimateCache(rdb)
	}

	// Cliente ML (com 2s de timeout)
	mlClient := client.NewMLClient(mlServiceURL, 2*time.Second)

	// Serviços
	rideService := service.NewRideService(userRepo, rideRepo, mlClient, estimateCache, eventPublisher)

	// Handlers HTTP & WebSockets
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userRepo)
	rideHandler := handler.NewRideHandler(rideService)
	authHandler := handler.NewAuthHandler(userRepo)
	driverHandler := handler.NewDriverHandler(geoRepo, eventPublisher, wsHub)
	wsHandler := handler.NewWSHandler(wsHub)

	// Roteador Chi com Auth, Rate Limiter, Redis GEO & WebSockets
	r := router.NewRouter(healthHandler, userHandler, rideHandler, authHandler, driverHandler, wsHandler)

	serverAddr := fmt.Sprintf(":%s", portStr)
	log.Printf("Servidor Routeforge API iniciado na porta %s (ML URL: %s)", portStr, mlServiceURL)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Erro no servidor HTTP: %v", err)
	}
}
