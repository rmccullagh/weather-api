package utils

import (
	"net/http/httptest"
	"testing"
)

func TestJSONResponse_Struct(t *testing.T) {
	type payload struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	obj := payload{Name: "Alice", Age: 30, Active: true}
	rr := httptest.NewRecorder()

	JSONResponse(rr, obj)

	expected := "{\n    \"name\": \"Alice\",\n    \"age\": 30,\n    \"active\": true\n}\n"
	if body := rr.Body.String(); body != expected {
		t.Fatalf("unexpected body:\n got: %q\nwant: %q", body, expected)
	}

	if rr.Code != 200 {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, 200)
	}
}

func TestJSONResponse_Nil(t *testing.T) {
	rr := httptest.NewRecorder()

	JSONResponse(rr, nil)

	expected := "null\n"
	if body := rr.Body.String(); body != expected {
		t.Fatalf("unexpected body for nil: got %q want %q", body, expected)
	}
}
