package models

type ForecastResponse struct {
	Properties struct {
		Periods []struct {
			Name          string `json:"name"`
			Temperature   int    `json:"temperature"`
			ShortForecast string `json:"shortForecast"`
		} `json:"periods"`
	} `json:"properties"`
}
