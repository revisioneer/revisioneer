package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRevisionsReturnsWithStatusOK(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestListRevisionsReturnsValidJSON(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	decoder := json.NewDecoder(response.Body)

	var deploy []Deployments
	err := decoder.Decode(&deploy)

	if err != nil {
		t.Fatalf("Decoding should pass: %v", err)
	}
}
