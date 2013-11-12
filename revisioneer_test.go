package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ClearDeployments() {
	Hd().Exec("DELETE FROM deployments")
}

func TestListRevisionsReturnsWithStatusOK(t *testing.T) {
	ClearDeployments()

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestListRevisionsReturnsValidJSON(t *testing.T) {
	ClearDeployments()

	var deploy Deployments = Deployments{Sha: "a", DeployedAt: time.Now()}
	_, _ = Hd().Save(&deploy)

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	decoder := json.NewDecoder(response.Body)

	var deployments []Deployments
	err := decoder.Decode(&deployments)

	if err != nil {
		t.Fatalf("Decoding should pass: %v", err)
	}
	if len(deployments) != 1 {
		t.Fatalf("Decoding should result in 1 element: %v", len(deployments))
	}
}
