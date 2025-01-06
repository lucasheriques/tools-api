package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	// Create a new application instance with rate limiting enabled
	app := &application{
		config: config{
			limiter: struct {
				rps     float64
				burst   int
				enabled bool
			}{
				rps:     2,
				burst:   4,
				enabled: true,
			},
		},
	}

	// Create a simple handler that always returns 200 OK
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap it with our rate limiter middleware
	handler := app.rateLimit(nextHandler)

	// Create a test server
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Helper function to make a request and return the status code
	makeRequest := func() int {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	// Test that we can make burst number of requests immediately
	for i := 0; i < app.config.limiter.burst; i++ {
		if status := makeRequest(); status != http.StatusOK {
			t.Errorf("Expected status %d for request %d, got %d", http.StatusOK, i+1, status)
		}
	}

	// The next request should be rate limited
	if status := makeRequest(); status != http.StatusTooManyRequests {
		t.Errorf("Expected status %d after burst, got %d", http.StatusTooManyRequests, status)
	}

	// Wait for a second to allow the rate limiter to recover
	time.Sleep(time.Second)

	// Should be able to make another request
	if status := makeRequest(); status != http.StatusOK {
		t.Errorf("Expected status %d after waiting, got %d", http.StatusOK, status)
	}
}
