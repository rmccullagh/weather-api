package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestGetForecast_ErrorConditions(t *testing.T) {
	tests := []struct {
		name       string
		transport  http.RoundTripper
		wantStatus int
		wantBody   string
	}{
		{
			name: "point network error",
			transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				if strings.HasPrefix(req.URL.Path, "/points/") {
					return nil, errors.New("network fail")
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"detail":"not used"}`)), Header: make(http.Header)}, nil
			}),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "network fail",
		},
		{
			name: "point non-200 with detail",
			transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				if strings.HasPrefix(req.URL.Path, "/points/") {
					return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"detail":"bad point"}`)), Header: make(http.Header)}, nil
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"detail":"not used"}`)), Header: make(http.Header)}, nil
			}),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "bad point",
		},
		{
			name: "forecast network error",
			transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/points/1,2":
					body := `{"properties":{"forecast":"https://api.weather.gov/forecast/1"}}`
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
				case "/forecast/1":
					return nil, errors.New("forecast network fail")
				default:
					return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"detail":"not found"}`)), Header: make(http.Header)}, nil
				}
			}),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "forecast network fail",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := http.DefaultTransport
			http.DefaultTransport = tc.transport
			defer func() { http.DefaultTransport = orig }()

			router := GetRouter()

			req := httptest.NewRequest("GET", "/v1/forecasts/1/2", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d", rr.Code, tc.wantStatus)
			}
			if !strings.Contains(rr.Body.String(), tc.wantBody) {
				t.Fatalf("body: expected to contain %q, got %s", tc.wantBody, rr.Body.String())
			}
		})
	}
}
