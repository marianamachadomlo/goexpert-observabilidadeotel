package servicea

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/marianamachado/observabilidade-otel/internal/cep"
	"github.com/marianamachado/observabilidade-otel/internal/otelsetup"
	"github.com/marianamachado/observabilidade-otel/internal/request"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	ServiceBURL string
	HTTPClient  *http.Client
}

func Run() {
	ctx := context.Background()

	serviceName := envOrDefault("OTEL_SERVICE_NAME", "service-a")
	collectorEndpoint := envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	port := envOrDefault("HTTP_PORT", ":8088")
	serviceBURL := envOrDefault("SERVICE_B_URL", "http://service-b:8081")

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

	server := &Server{
		ServiceBURL: serviceBURL,
		HTTPClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   30 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("POST /", otelhttp.NewHandler(http.HandlerFunc(server.handleWeather), "service-a.handleWeather"))

	log.Printf("service-a listening on %s", port)
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

	forwardBody, err := json.Marshal(map[string]string{"cep": zipcode})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, s.ServiceBURL, bytes.NewReader(forwardBody))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		http.Error(w, "service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
