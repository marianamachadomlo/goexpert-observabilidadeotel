package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Location struct {
	City string
	UF   string
}

type ViaCEPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewViaCEPClient(httpClient *http.Client) *ViaCEPClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ViaCEPClient{
		BaseURL:    "https://viacep.com.br/ws",
		HTTPClient: httpClient,
	}
}

type viaCEPResponse struct {
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"`
}

func (c *ViaCEPClient) Lookup(ctx context.Context, zipcode string) (*Location, error) {
	url := fmt.Sprintf("%s/%s/json/", c.BaseURL, zipcode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request viacep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viacep returned status %d", resp.StatusCode)
	}

	var payload viaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode viacep response: %w", err)
	}

	if payload.Erro || payload.Localidade == "" {
		return nil, ErrNotFound
	}

	return &Location{
		City: payload.Localidade,
		UF:   payload.UF,
	}, nil
}
