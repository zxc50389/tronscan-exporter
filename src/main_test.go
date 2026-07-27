package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchTPS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"blockHeight":123,"currentTps":45.6,"maxTps":78.9},"type":"ok"}`))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	data, err := fetchTPS(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Data.BlockHeight != 123 {
		t.Errorf("BlockHeight = %v, want 123", data.Data.BlockHeight)
	}
	if data.Data.CurrentTps != 45.6 {
		t.Errorf("CurrentTps = %v, want 45.6", data.Data.CurrentTps)
	}
	if data.Data.MaxTps != 78.9 {
		t.Errorf("MaxTps = %v, want 78.9", data.Data.MaxTps)
	}
}

func TestFetchTPSNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	if _, err := fetchTPS(context.Background(), client, srv.URL); err == nil {
		t.Fatal("expected an error for a non-200 response, got nil")
	}
}

func TestFetchTPSInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	if _, err := fetchTPS(context.Background(), client, srv.URL); err == nil {
		t.Fatal("expected an error for invalid JSON, got nil")
	}
}
