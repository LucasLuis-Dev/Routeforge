package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
)

type mlClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient(baseURL string, timeout time.Duration) domain.PredictionClient {
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	return &mlClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *mlClient) Predict(ctx context.Context, req *domain.PredictionRequest) (*domain.PredictionResponse, error) {
	url := fmt.Sprintf("%s/predict", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição ML: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP para ML: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha de comunicação com microsserviço de ML (timeout/erro): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("microsserviço de ML respondeu com status HTTP %d", resp.StatusCode)
	}

	var predResp domain.PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON do ML: %w", err)
	}

	return &predResp, nil
}
