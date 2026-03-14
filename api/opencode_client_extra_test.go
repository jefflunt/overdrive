package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenCodeClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/global/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &OpenCodeClient{BaseURL: server.URL}
	if err := client.Health(); err != nil {
		t.Errorf("Health() failed: %v", err)
	}
}

func TestOpenCodeClient_CreateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session" && r.Method == "POST" {
			json.NewEncoder(w).Encode(OpenCodeSession{ID: "s1", Title: "Test"})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &OpenCodeClient{BaseURL: server.URL}
	id, err := client.CreateSession("Test")
	if err != nil {
		t.Errorf("CreateSession() failed: %v", err)
	}
	if id != "s1" {
		t.Errorf("Expected id s1, got %s", id)
	}
}

func TestOpenCodeClient_IsGenerating(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session/status" {
			json.NewEncoder(w).Encode(map[string]string{"s1": "generating", "s2": "idle"})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &OpenCodeClient{BaseURL: server.URL}
	gen, err := client.IsGenerating("s1")
	if err != nil || !gen {
		t.Errorf("IsGenerating(s1) failed: %v, %v", err, gen)
	}

	gen, err = client.IsGenerating("s2")
	if err != nil || gen {
		t.Errorf("IsGenerating(s2) failed: %v, %v", err, gen)
	}
}
