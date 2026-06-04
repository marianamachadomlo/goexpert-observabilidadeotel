package temperature

import "math"

type Response struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func FromCelsius(city string, celsius float64) Response {
	return Response{
		City:  city,
		TempC: round(celsius, 1),
		TempF: round(celsius*1.8+32, 1),
		TempK: round(celsius+273, 2),
	}
}

func round(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}
