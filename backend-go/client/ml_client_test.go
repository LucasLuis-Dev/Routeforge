package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestMLClient_CircuitBreaker_Tripping(t *testing.T) {
	// Servidor HTTP simulado que falha com HTTP 500
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer failServer.Close()

	cli := NewMLClient(failServer.URL, 1*time.Second)

	req := &domain.PredictionRequest{
		DistanceKM: 5.0,
		HourOfDay:  14,
		DayOfWeek:  2,
	}

	// Executa 3 chamadas que falham -> devem disparar o Circuit Breaker para OPEN
	for i := 0; i < 3; i++ {
		_, err := cli.Predict(context.Background(), req)
		assert.Error(t, err)
	}

	// A 4ª chamada deve falhar instantaneamente com gobreaker.ErrOpenState sem nem tentar bater na rede
	_, err := cli.Predict(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, gobreaker.ErrOpenState, err, "Circuit Breaker deve estar no estado OPEN após 50%+ de falhas")
}
