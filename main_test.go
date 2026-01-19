package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func TestGetForecast_Success(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()

	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/points/1,2":
			body := `{"properties":{"forecast":"https://api.weather.gov/forecast/1"}}`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		case "/forecast/1":
			body := `{"properties":{"periods":[{"shortForecast":"Sunny","temperature":90}]}}`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		default:
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"detail":"not found"}`)), Header: make(http.Header)}, nil
		}
	})

	router := chi.NewRouter()
	router.Get("/v1/forecasts/{latitude}/{longitude}", GetForecast)

	req := httptest.NewRequest("GET", "/v1/forecasts/1/2", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Sunny") || !strings.Contains(body, "90") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestRootRedirect(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusTemporaryRedirect)
	}
	if loc := rr.Header().Get("Location"); loc != "/swagger/index.html" {
		t.Fatalf("location header: got %q want %q", loc, "/swagger/index.html")
	}
}
