# Observabilidade OTEL — Clima por CEP (Go Expert)

Sistema distribuído em Go com dois microsserviços que consultam o clima de uma cidade a partir do CEP, com rastreamento distribuído via OpenTelemetry e Zipkin.

## Arquitetura

```
Cliente → Serviço A (input) → Serviço B (orquestração) → ViaCEP + WeatherAPI
                ↓                      ↓
           OTEL Collector ←───────────┘
                ↓
              Zipkin
```

- **Serviço A**: valida o CEP e encaminha a requisição ao Serviço B.
- **Serviço B**: busca a cidade (ViaCEP), consulta a temperatura (WeatherAPI) e retorna Celsius, Fahrenheit e Kelvin.
- **OTEL Collector**: recebe os traços via OTLP e envia ao Zipkin.

## Pré-requisitos

- Docker e Docker Compose
- Chave da [WeatherAPI](https://www.weatherapi.com/)

## Configuração

Copie o arquivo de exemplo e informe sua chave:

```bash
cp .env.sample .env
```

Edite `.env` e defina `WEATHER_API_KEY`.

## Executar

```bash
docker compose up --build
```

Serviços disponíveis:

| Serviço         | URL                    |
|-----------------|------------------------|
| Serviço A       | http://localhost:8088  |
| Serviço B       | http://localhost:8081  |
| Zipkin          | http://localhost:9411  |
| OTEL Collector  | localhost:4317 (gRPC)  |

## Requisição POST no Serviço A

```bash
curl -X POST http://localhost:8088 \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'
```

Resposta de sucesso (200):

```json
{
  "city": "Vitória",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

Erros:

| Código | Mensagem              | Causa                          |
|--------|-----------------------|--------------------------------|
| 422    | invalid zipcode       | CEP inválido ou não é string   |
| 404    | can not find zipcode  | CEP não encontrado             |

Exemplo de CEP inválido:

```bash
curl -X POST http://localhost:8088 \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

## Visualizar traços no Zipkin

1. Acesse [http://localhost:9411](http://localhost:9411).
2. Clique em **Run Query** (ou use o botão de busca).
3. Selecione um trace para ver o fluxo completo: **Request → Serviço A → Serviço B**.
4. Dentro do Serviço B, os spans manuais **Busca de CEP** e **Busca de Temperatura** medem as chamadas às APIs externas.