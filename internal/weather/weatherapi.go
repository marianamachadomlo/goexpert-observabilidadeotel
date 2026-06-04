package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		APIKey:     apiKey,
		BaseURL:    "https://api.weatherapi.com/v1",
		HTTPClient: httpClient,
	}
}

type currentResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func (c *Client) CurrentCelsius(ctx context.Context, city, uf string) (float64, error) {
	query := fmt.Sprintf("%s, %s, Brazil", city, uf)
	endpoint := fmt.Sprintf("%s/current.json?key=%s&q=%s", c.BaseURL, url.QueryEscape(c.APIKey), url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request weatherapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi returned status %d", resp.StatusCode)
	}

	var payload currentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, fmt.Errorf("decode weatherapi response: %w", err)
	}

	return payload.Current.TempC, nil
}
