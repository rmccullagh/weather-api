package models

import "testing"

func TestMapCharacterizationFromTemp(t *testing.T) {
	tests := []struct {
		name string
		temp int
		want Characterization
	}{
		{"hot high", 90, Hot},
		{"hot boundary", 85, Hot},
		{"moderate low", 60, Moderate},
		{"moderate mid", 70, Moderate},
		{"moderate high", 75, Moderate},
		{"cold high", 50, Cold},
		{"cold low", 30, Cold},
		{"unknown low", 55, Unknown},
		{"unknown high", 76, Unknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MapCharacterizationFromTemp(tc.temp); got != tc.want {
				t.Fatalf("temp=%d: got %q want %q", tc.temp, got, tc.want)
			}
		})
	}
}

func TestNewForecastFromUpstream(t *testing.T) {
	tests := []struct {
		name          string
		temp          int
		shortForecast string
		wantTemp      int
		wantForecast  string
		wantChar      Characterization
	}{
		{"basic moderate", 70, "Partly Sunny", 70, "Partly Sunny", Moderate},
		{"hot case", 90, "Hot Day", 90, "Hot Day", Hot},
		{"cold case", 40, "Cold Morning", 40, "Cold Morning", Cold},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fr := &ForecastResponse{
				Properties: struct {
					Periods []struct {
						Name          string `json:"name"`
						Temperature   int    `json:"temperature"`
						ShortForecast string `json:"shortForecast"`
					} `json:"periods"`
				}{
					Periods: []struct {
						Name          string `json:"name"`
						Temperature   int    `json:"temperature"`
						ShortForecast string `json:"shortForecast"`
					}{
						{
							ShortForecast: tc.shortForecast,
							Temperature:   tc.temp,
						},
					},
				},
			}

			got := NewForecastFromUpstream(fr)
			if got == nil {
				t.Fatal("NewForecastFromUpstream returned nil")
			}

			if got.ForecastDaily != tc.wantForecast {
				t.Fatalf("ForecastDaily: got %q want %q", got.ForecastDaily, tc.wantForecast)
			}

			if got.Temperature != tc.wantTemp {
				t.Fatalf("Temperature: got %d want %d", got.Temperature, tc.wantTemp)
			}

			if got.Characterization != tc.wantChar {
				t.Fatalf("Characterization: got %q want %q", got.Characterization, tc.wantChar)
			}
		})
	}
}
