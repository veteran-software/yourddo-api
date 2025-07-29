package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	workers := 5
	maxRetries := 3
	pool := NewWorkerPool(workers, maxRetries)

	if pool.workers != workers {
		t.Errorf("NewWorkerPool() workers = %v, want %v", pool.workers, workers)
	}
	if pool.maxRetries != maxRetries {
		t.Errorf("NewWorkerPool() maxRetries = %v, want %v", pool.maxRetries, maxRetries)
	}
	if pool.client == nil {
		t.Error("NewWorkerPool() client is nil")
	}
}

func TestWorkerPoolProcessURLs(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, err := w.Write([]byte(`<Status><name>TestServer</name></Status>`))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		urls    []string
		workers int
		wantLen int
		wantErr bool
		timeout time.Duration
	}{
		{
			name:    "single URL",
			urls:    []string{server.URL},
			workers: 1,
			wantLen: 1,
			wantErr: false,
			timeout: 5 * time.Second,
		},
		{
			name:    "multiple URLs",
			urls:    []string{server.URL, server.URL, server.URL},
			workers: 2,
			wantLen: 3,
			wantErr: false,
			timeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTests(t, tt)
		})
	}
}

func runTests(t *testing.T, tt struct {
	name    string
	urls    []string
	workers int
	wantLen int
	wantErr bool
	timeout time.Duration
}) {
	ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
	defer cancel()

	pool := NewWorkerPool(tt.workers, 3)
	results := pool.ProcessURLs(ctx, tt.urls)

	count := 0
	for result := range results {
		count++
		if result.Error != nil && !tt.wantErr {
			t.Errorf("ProcessURLs() got unexpected error: %v", result.Error)
		}
		if result.Status == nil && !tt.wantErr {
			t.Error("ProcessURLs() got nil Status for successful request")
		}
	}

	if count != tt.wantLen {
		t.Errorf("ProcessURLs() processed %v URLs, want %v", count, tt.wantLen)
	}
}

func TestWorkerPoolFetchStatus(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, err := w.Write([]byte(`<Status><name>TestServer</name></Status>`))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     server.URL,
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "http://invalid.url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(1, 1)
			ctx := context.Background()

			status, err := pool.fetchStatus(ctx, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && status == nil {
				t.Error("fetchStatus() returned nil status for successful request")
			}
		})
	}
}
