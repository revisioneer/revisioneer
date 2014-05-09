package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var deploymentsController *DeploymentsController

func init() {
	_db = Setup()
	deploymentsController = NewDeploymentsController(_db)
}

func ClearDeployments() {
	_db.Query("DELETE FROM messages").Run()
	_db.Query("DELETE FROM deployments").Run()
	_db.Query("DELETE FROM projects").Run()
}

func CreateTestProject(apiToken string) Project {
	if apiToken == "" {
		apiToken = "test"
	}

	var project Project = Project{Name: "Test", ApiToken: apiToken}
	project.Store(_db)
	return project
}

func CreateTestDeployment(project Project, sha string) Deployment {
	var deployedAt time.Time = time.Now()
	var deploy Deployment = Deployment{Sha: sha, DeployedAt: deployedAt, ProjectId: int(project.Id)}
	deploy.Store(_db)
	return deploy
}

func TestCreateDeploymentReturnsCreatedRevision(t *testing.T) {
	ClearDeployments()

	project := CreateTestProject("")

	request, _ := http.NewRequest("POST", "/deployments", strings.NewReader(`{"sha":"asd","messages": ["A Message"]}`))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.CreateDeployment(response, request, project)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}

	decoder := json.NewDecoder(response.Body)
	var newDeploy Deployment
	_ = decoder.Decode(&newDeploy)

	if newDeploy.Sha != "asd" {
		t.Fatalf("Did not read proper SHA: %v", newDeploy.Sha)
	}

	var deployments []Deployment
	err := _db.Query(`SELECT * FROM deployments ORDER BY deployed_at`).Rows(&deployments)
	if err != nil {
		t.Fatalf("Unable to read from PostgreSQL: %v", err)
	}
	if len(deployments) != 1 {
		t.Fatalf("More than 1 entry created: %d", len(deployments))
	}
}

func TestVerifyDeploymentWithUnknownRevision(t *testing.T) {
	ClearDeployments()

	project := CreateTestProject("")
	request, _ := http.NewRequest("POST", "/deployments/revision/verify", strings.NewReader(""))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.VerifyDeployment(response, request, project, map[string]string{"sha": "revision"})

	if response.Code != http.StatusNotFound {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestVerifyDeployment(t *testing.T) {
	ClearDeployments()

	project := CreateTestProject("")
	deployment := CreateTestDeployment(project, "revision")

	request, _ := http.NewRequest("POST", "/deployments/revision/verify", strings.NewReader(""))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.VerifyDeployment(response, request, project, map[string]string{"sha": "revision"})

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}

	_db.Query(`SELECT * FROM deployments WHERE id = $1`, deployment.Id).Rows(&deployment)

	if !deployment.Verified {
		t.Fatalf(`Deployment should have been verified: %v`, deployment)
	}
	if deployment.VerifiedAt.IsZero() {
		t.Fatalf("Deployment should have been verified_at")
	}
}

func TestListDeploymentsReturnsWithStatusOK(t *testing.T) {
	project := CreateTestProject("")

	request, _ := http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.ListDeployments(response, request, project)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}

func TestRevisionsAreScopedByApiToken(t *testing.T) {
	projectA := CreateTestProject("testA")
	projectB := CreateTestProject("testB")

	revA := CreateTestDeployment(projectA, "a")
	revB := CreateTestDeployment(projectB, "b")

	request, _ := http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", projectA.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.ListDeployments(response, request, projectA)

	decoder := json.NewDecoder(response.Body)

	var deploymentsA []Deployment
	_ = decoder.Decode(&deploymentsA)
	if deploymentsA[0].Sha != revA.Sha || len(deploymentsA) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsA)
	}

	request, _ = http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", projectB.ApiToken)
	response = httptest.NewRecorder()

	deploymentsController.ListDeployments(response, request, projectB)

	decoder = json.NewDecoder(response.Body)

	var deploymentsB []Deployment
	_ = decoder.Decode(&deploymentsB)
	if deploymentsB[0].Sha != revB.Sha || len(deploymentsB) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsB)
	}
}

func TestListDeploymentsReturnsValidJSON(t *testing.T) {
	project := CreateTestProject("")

	var deploy Deployment = CreateTestDeployment(project, "test")

	request, _ := http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	deploymentsController.ListDeployments(response, request, project)

	decoder := json.NewDecoder(response.Body)

	var deployments []Deployment
	err := decoder.Decode(&deployments)

	if err != nil {
		t.Fatalf("Decoding should pass: %v", err)
	}
	if len(deployments) != 1 || deploy.Sha != "test" {
		t.Fatalf("Decoding failed: %v", deployments)
	}
}
