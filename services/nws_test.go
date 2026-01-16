package services

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rmccullagh/weather-api/models"
)

func TestDoHTTPGet_Success(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, `{"properties":{"forecast":"`+ts.URL+`/forecast"}}`)
	}))
	defer ts.Close()

	got, err := doHTTPGet[pointResponse](ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Properties.Forecast != ts.URL+"/forecast" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestDoHTTPGet_Non200_WithErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		io.WriteString(w, `{"detail":"bad request happened"}`)
	}))
	defer ts.Close()

	_, err := doHTTPGet[pointResponse](ts.URL)
	if err == nil || !strings.Contains(err.Error(), "bad request happened") {
		t.Fatalf("expected error containing detail, got: %v", err)
	}
}

func TestDoHTTPGet_Non200_NonJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		io.WriteString(w, `internal server error`)
	}))
	defer ts.Close()

	_, err := doHTTPGet[pointResponse](ts.URL)
	if err == nil || !strings.Contains(err.Error(), "non 200 response from upstream") {
		t.Fatalf("expected non-200 non-json error, got: %v", err)
	}
}

type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network fail")
}

func TestDoHTTPGet_NetworkError(t *testing.T) {
	orig := http.DefaultTransport
	http.DefaultTransport = errRoundTripper{}
	defer func() { http.DefaultTransport = orig }()

	_, err := doHTTPGet[pointResponse]("http://example.invalid")
	if err == nil || !strings.Contains(err.Error(), "network fail") {
		t.Fatalf("expected network error, got: %v", err)
	}
}

func TestNwsAPI_GetForecast_Success(t *testing.T) {
	origTransport := http.DefaultTransport
	// restore at end
	defer func() { http.DefaultTransport = origTransport }()

	// Create a transport that responds to the two expected paths.
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/points/1,2":
			body := `{"properties":{"forecast":"https://api.weather.gov/forecast/1"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		case "/forecast/1":
			// Return ForecastResponse JSON expected by models.NewForecastFromUpstream
			body := `{
                "properties": {
                    "periods": [
                        {
                            "shortForecast": "Sunny",
                            "temperature": 90
                        }
                    ]
                }
            }`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		default:
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		}
	})

	c := NewClient()
	f, err := c.GetForecast("1", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected forecast, got nil")
	}
	if f.ForecastDaily != "Sunny" || f.Temperature != 90 || f.Characterization != models.Hot {
		t.Fatalf("unexpected forecast: %#v", f)
	}
}

func TestNwsAPI_GetForecast_PointNetworkError(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()

	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if strings.HasPrefix(req.URL.Path, "/points/") {
			return nil, errors.New("network fail")
		}
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader(`{"detail":"not used"}`)),
			Header:     make(http.Header),
		}, nil
	})

	c := NewClient()
	_, err := c.GetForecast("1", "2")
	if err == nil || !strings.Contains(err.Error(), "network fail") {
		t.Fatalf("expected network error from points request, got: %v", err)
	}
}

func TestNwsAPI_GetForecast_PointNon200_WithDetail(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()

	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/points/1,2" {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"bad point"}`)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader(`{"detail":"not used"}`)),
			Header:     make(http.Header),
		}, nil
	})

	c := NewClient()
	_, err := c.GetForecast("1", "2")
	if err == nil || !strings.Contains(err.Error(), "bad point") {
		t.Fatalf("expected detail error from points request, got: %v", err)
	}
}

func TestNwsAPI_GetForecast_ForecastNon200_NonJSON(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()

	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/points/1,2":
			body := `{"properties":{"forecast":"https://api.weather.gov/forecast/1"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		case "/forecast/1":
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("internal server error")),
				Header:     make(http.Header),
			}, nil
		default:
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		}
	})

	c := NewClient()
	_, err := c.GetForecast("1", "2")
	if err == nil || !strings.Contains(err.Error(), "non 200 response from upstream") {
		t.Fatalf("expected non-200 non-json error from forecast request, got: %v", err)
	}
}

func TestNwsAPI_GetForecast_ForecastNon200_WithDetail(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()

	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/points/1,2":
			body := `{"properties":{"forecast":"https://api.weather.gov/forecast/1"}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		case "/forecast/1":
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		default:
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"not found"}`)),
				Header:     make(http.Header),
			}, nil
		}
	})

	c := NewClient()
	_, err := c.GetForecast("1", "2")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected detail error from forecast request, got: %v", err)
	}
}

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
