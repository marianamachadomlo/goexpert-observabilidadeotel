package serviceb

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/marianamachado/observabilidade-otel/internal/cep"
	"github.com/marianamachado/observabilidade-otel/internal/otelsetup"
	"github.com/marianamachado/observabilidade-otel/internal/request"
	"github.com/marianamachado/observabilidade-otel/internal/service"
	"github.com/marianamachado/observabilidade-otel/internal/weather"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	WeatherService *service.WeatherService
}

func Run() {
	ctx := context.Background()

	serviceName := envOrDefault("OTEL_SERVICE_NAME", "service-b")
	collectorEndpoint := envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	port := envOrDefault("HTTP_PORT", ":8081")
	weatherAPIKey := os.Getenv("WEATHER_API_KEY")
	if weatherAPIKey == "" {
		log.Fatal("WEATHER_API_KEY is required")
	}

	shutdown, err := otelsetup.InitTracer(ctx, serviceName, collectorEndpoint)
	if err != nil {
		log.Fatalf("init tracer: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown tracer: %v", err)
		}
	}()

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   30 * time.Second,
	}

	weatherService := &service.WeatherService{
		CEPClient:     cep.NewViaCEPClient(httpClient),
		WeatherClient: weather.NewClient(weatherAPIKey, httpClient),
		TracerName:    "service-b",
	}

	server := &Server{WeatherService: weatherService}

	mux := http.NewServeMux()
	mux.Handle("POST /", otelhttp.NewHandler(http.HandlerFunc(server.handleWeather), "service-b.handleWeather"))

	log.Printf("service-b listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func (s *Server) handleWeather(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	var payload request.ZipcodePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	zipcode, err := request.ParseZipcode(payload.CEP)
	if err != nil || !cep.IsValid(zipcode) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	result, err := s.WeatherService.GetTemperatureByZipcode(r.Context(), zipcode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidZipcode):
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		case errors.Is(err, cep.ErrNotFound):
			http.Error(w, "can not find zipcode", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
