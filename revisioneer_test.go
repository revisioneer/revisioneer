package main

import (
	"encoding/json"
	"github.com/eaigner/hood"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func ClearDeployments() *hood.Hood {
	hd := Hd()
	hd.Exec("DELETE FROM deployments")
	hd.Exec("DELETE FROM projects")
	return hd
}

func CreateTestProject(hd *hood.Hood, apiToken string) Projects {
	if apiToken == "" {
		apiToken = "test"
	}

	var project Projects = Projects{Name: "Test", ApiToken: apiToken}
	hd.Save(&project)
	return project
}

func CreateTestRevision(hd *hood.Hood, project Projects, sha string) Deployments {
	var deployedAt time.Time = time.Now()
	var deploy Deployments = Deployments{Sha: sha, DeployedAt: deployedAt, ProjectId: int(project.Id)}
	_, _ = hd.Save(&deploy)
	return deploy
}

func TestCreateRevisionReturnsCreatedRevision(t *testing.T) {
	hd := ClearDeployments()
	defer hd.Db.Close()

	project := CreateTestProject(hd, "")

	request, _ := http.NewRequest("POST", "/revisions", strings.NewReader("{\"sha\":\"asd\"}"))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	CreateRevision(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}

	var deployments []Deployments
	err := hd.OrderBy("deployed_at").Find(&deployments)
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
	hd := ClearDeployments()
	defer hd.Db.Close()
	project := CreateTestProject(hd, "")

	request, _ := http.NewRequest("GET", "/revisions", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestRevisionsAreScopedByApiToken(t *testing.T) {
	hd := ClearDeployments()
	defer hd.Db.Close()

	projectA := CreateTestProject(hd, "testA")
	projectB := CreateTestProject(hd, "testB")

	revA := CreateTestRevision(hd, projectA, "a")
	revB := CreateTestRevision(hd, projectB, "b")

	request, _ := http.NewRequest("GET", "/revisions", nil)
	request.Header.Set("API-TOKEN", projectA.ApiToken)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	decoder := json.NewDecoder(response.Body)

	var deploymentsA []Deployments
	_ = decoder.Decode(&deploymentsA)
	if deploymentsA[0].Sha != revA.Sha || len(deploymentsA) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsA)
	}

	request, _ = http.NewRequest("GET", "/revisions", nil)
	request.Header.Set("API-TOKEN", projectB.ApiToken)
	response = httptest.NewRecorder()

	ListRevisions(response, request)

	decoder = json.NewDecoder(response.Body)

	var deploymentsB []Deployments
	_ = decoder.Decode(&deploymentsB)
	if deploymentsB[0].Sha != revB.Sha || len(deploymentsB) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsB)
	}
}

func TestListRevisionsReturnsValidJSON(t *testing.T) {
	hd := ClearDeployments()
	defer hd.Db.Close()
	project := CreateTestProject(hd, "")

	var deploy Deployments = CreateTestRevision(hd, project, "test")

	request, _ := http.NewRequest("GET", "/revisions", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	ListRevisions(response, request)

	decoder := json.NewDecoder(response.Body)

	var deployments []Deployments
	err := decoder.Decode(&deployments)

	if err != nil {
		t.Fatalf("Decoding should pass: %v", err)
	}
	if len(deployments) != 1 || deploy.Sha != "test" {
		t.Fatalf("Decoding failed: %v", deployments)
	}
}
