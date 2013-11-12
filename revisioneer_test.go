package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
  "strings"
)

func ClearDeployments() {
	Hd().Exec("DELETE FROM deployments")
}

func TestCreateRevisionReturnsCreatedRevision(t *testing.T) {
  ClearDeployments()

  request, _ := http.NewRequest("POST", "/revisions", strings.NewReader("{\"sha\":\"asd\"}"))
  response := httptest.NewRecorder()

  CreateRevision(response, request)

  if response.Code != http.StatusOK {
    t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
  }

  var deployments []Deployments
  err := Hd().OrderBy("deployed_at").Find(&deployments)
  if err != nil {
    t.Fatalf("Unable to read from PostgreSQL: %v", err)
  }
  if len(deployments) != 1 {
    t.Fatalf("More than 1 entry created: %d", len(deployments))
  }

  var newDeploy Deployments = deployments[0]
  if newDeploy.Sha != "asd" {
    t.Fatalf("Did not read proper SHA: %v", newDeploy.Sha)
  }
}

func TestListRevisionsReturnsWithStatusOK(t *testing.T) {
	ClearDeployments()

	request, _ := http.NewRequest("GET", "/revisions", nil)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestListRevisionsReturnsValidJSON(t *testing.T) {
	ClearDeployments()

  var deployedAt time.Time = time.Now()
	var deploy Deployments = Deployments{Sha: "a", DeployedAt: deployedAt}
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
	if len(deployments) != 1 || deploy.DeployedAt != deployedAt || deploy.Sha != "a" {
		t.Fatalf("Decoding failed: %v", deployments)
	}
}
