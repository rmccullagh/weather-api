package models

type Characterization string

const (
	Hot      Characterization = "hot"
	Cold     Characterization = "cold"
	Moderate Characterization = "moderate"
	Unknown  Characterization = "unknown "
)

type Forecast struct {
	ForecastDaily    string           `json:"forecast_daily"`
	Characterization Characterization `json:"temperature_characterization"`
	Temperature      int              `json:"temperature"`
}

func MapCharacterizationFromTemp(temp int) Characterization {
	if temp >= 85 {
		return Hot
	}

	if temp >= 60 && temp <= 75 {
		return Moderate
	}

	if temp <= 50 {
		return Cold
	}

	return Unknown
}

func NewForecastFromUpstream(upstream *ForecastResponse) *Forecast {
	return &Forecast{
		ForecastDaily:    upstream.Properties.Periods[0].ShortForecast,
		Characterization: MapCharacterizationFromTemp(upstream.Properties.Periods[0].Temperature),
		Temperature:      upstream.Properties.Periods[0].Temperature,
	}
}
