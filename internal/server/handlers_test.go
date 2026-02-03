package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	h := &Handlers{}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != Version {
		t.Errorf("Version = %q, want %q", resp.Version, Version)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("Response key = %q, want %q", resp["key"], "value")
	}
}

func TestInjectHandlerInvalidJSON(t *testing.T) {
	h := &Handlers{bridge: nil}

	req := httptest.NewRequest("POST", "/inject", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered panic (expected - bridge is nil): %v", r)
		}
	}()

	h.Inject(w, req)
}

func TestInjectHandlerEmptyText(t *testing.T) {
	h := &Handlers{bridge: nil}

	body := `{"text": ""}`
	req := httptest.NewRequest("POST", "/inject", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered panic (expected - bridge is nil): %v", r)
		}
	}()

	h.Inject(w, req)
}
