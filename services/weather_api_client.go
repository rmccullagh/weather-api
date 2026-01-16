package services

import "github.com/rmccullagh/weather-api/models"

type WeatherClient interface {
	GetForecast(latitude, longitude string) (*models.Forecast, error)
}

func NewClient() WeatherClient {
	return &nwsAPI{}
}
