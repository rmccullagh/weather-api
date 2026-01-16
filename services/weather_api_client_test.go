package services

import "testing"

func TestNewClient_NotNil(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClient_ConcreteType(t *testing.T) {
	c := NewClient()
	if _, ok := c.(*nwsAPI); !ok {
		t.Fatalf("expected *nwsAPI concrete type, got %T", c)
	}
}
