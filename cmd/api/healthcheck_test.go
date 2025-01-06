package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tools.lucasfaria.dev/internal/assert"
)

func TestHealthcheckHandler(t *testing.T) {
	// Create a new application instance
	app := &application{
		config: config{
			env: "testing",
		},
	}

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)

	// Create a new recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the healthcheck handler directly
	app.healthcheckHandler(rr, req)

	// Check the status code
	assert.Equal(t, rr.Code, http.StatusOK)

	// Check the Content-Type header
	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")

	// Parse and check the response body
	var response struct {
		Status     string `json:"status"`
		SystemInfo struct {
			Environment string `json:"environment"`
			Version     string `json:"version"`
		} `json:"system_info"`
	}

	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response fields
	assert.Equal(t, response.Status, "available")
	assert.Equal(t, response.SystemInfo.Environment, "testing")
	assert.Equal(t, response.SystemInfo.Version, version)
}
