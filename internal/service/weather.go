package service

import (
	"context"
	"errors"

	"github.com/marianamachado/observabilidade-otel/internal/cep"
	"github.com/marianamachado/observabilidade-otel/internal/temperature"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type LocationLookup interface {
	Lookup(ctx context.Context, zipcode string) (*cep.Location, error)
}

type WeatherLookup interface {
	CurrentCelsius(ctx context.Context, city, uf string) (float64, error)
}

type WeatherService struct {
	CEPClient     LocationLookup
	WeatherClient WeatherLookup
	TracerName    string
}

func (s *WeatherService) GetTemperatureByZipcode(ctx context.Context, zipcode string) (temperature.Response, error) {
	if !cep.IsValid(zipcode) {
		return temperature.Response{}, ErrInvalidZipcode
	}

	tracer := otel.Tracer(s.TracerName)

	ctx, span := tracer.Start(ctx, "Busca de CEP")
	span.SetAttributes(attribute.String("cep", zipcode))
	location, err := s.CEPClient.Lookup(ctx, zipcode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		if errors.Is(err, cep.ErrNotFound) {
			return temperature.Response{}, cep.ErrNotFound
		}
		return temperature.Response{}, err
	}
	span.End()

	ctx, span = tracer.Start(ctx, "Busca de Temperatura")
	span.SetAttributes(
		attribute.String("city", location.City),
		attribute.String("uf", location.UF),
	)
	celsius, err := s.WeatherClient.CurrentCelsius(ctx, location.City, location.UF)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return temperature.Response{}, err
	}
	span.End()

	return temperature.FromCelsius(location.City, celsius), nil
}
