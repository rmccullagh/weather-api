package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rmccullagh/weather-api/models"
)

const baseURL = "https://api.weather.gov"

type nwsAPI struct{}

type pointResponse struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

type errorResponse struct {
	Detail string `json:"detail"`
}

func doHTTPGet[T any](endpoint string) (*T, error) {
	resp, err := http.DefaultClient.Get(endpoint)

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		// try to get the error
		var errorResponse errorResponse

		err = json.Unmarshal(body, &errorResponse)

		if err == nil {
			return nil, errors.New(errorResponse.Detail)
		} else {
			return nil, errors.New("non 200 response from upstream")
		}
	}

	var model T

	err = json.Unmarshal(body, &model)

	if err != nil {
		return nil, err
	}

	return &model, nil
}

// See https://www.weather.gov/documentation/services-web-api
func (n *nwsAPI) GetForecast(latitude, longitude string) (*models.Forecast, error) {
	point, err := doHTTPGet[pointResponse](baseURL + fmt.Sprintf("/points/%s,%s", latitude, longitude))

	if err != nil {
		return nil, err
	}

	forecast, err := doHTTPGet[models.ForecastResponse](point.Properties.Forecast)

	if err != nil {
		return nil, err
	}

	// TODO: cache the response in memory to reduce latency (key: URL, value: JSON blob)

	return models.NewForecastFromUpstream(forecast), nil
}
