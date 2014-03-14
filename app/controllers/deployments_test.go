package controllers

import (
	. "../models"
	"encoding/json"
	_ "github.com/eaigner/hood"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	base = &Base{}
	base.Setup()
}

func ClearDeployments() {
	base.Hd.Exec("DELETE FROM messages")
	base.Hd.Exec("DELETE FROM deployments")
	base.Hd.Exec("DELETE FROM projects")
}

func CreateTestProject(apiToken string) Projects {
	if apiToken == "" {
		apiToken = "test"
	}

	var project Projects = Projects{Name: "Test", ApiToken: apiToken}
	base.Hd.Save(&project)
	return project
}

func CreateTestDeployment(project Projects, sha string) Deployments {
	var deployedAt time.Time = time.Now()
	var deploy Deployments = Deployments{Sha: sha, DeployedAt: deployedAt, ProjectId: int(project.Id)}
	_, _ = base.Hd.Save(&deploy)
	return deploy
}

func TestCreateDeploymentReturnsCreatedRevision(t *testing.T) {
	ClearDeployments()

	project := CreateTestProject("")

	request, _ := http.NewRequest("POST", "/deployments", strings.NewReader("{\"sha\":\"asd\"}"))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	base.CreateDeployment(response, request, project)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}

	var deployments []Deployments
	err := base.Hd.OrderBy("deployed_at").Find(&deployments)
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

func TestVerifyDeploymentWithUnknownRevision(t *testing.T) {
	ClearDeployments()

	project := CreateTestProject("")
	request, _ := http.NewRequest("POST", "/deployments/revision/verify", strings.NewReader(""))
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	base.VerifyDeployment(response, request, project, map[string]string{"sha": "revision"})

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

	base.VerifyDeployment(response, request, project, map[string]string{"sha": "revision"})

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}

	var deployments []Deployments
	base.Hd.Where("id", "=", deployment.Id).Find(&deployments)
	if len(deployments) != 1 {
		t.Fatalf("Wrong number of deployments")
	}

	deployment = deployments[0]
	if !deployment.Verified {
		t.Fatalf("Deployment should have been verified")
	}
}

func TestListDeploymentsReturnsWithStatusOK(t *testing.T) {
	project := CreateTestProject("")

	request, _ := http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	base.ListDeployments(response, request, project)

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

	base.ListDeployments(response, request, projectA)

	decoder := json.NewDecoder(response.Body)

	var deploymentsA []Deployments
	_ = decoder.Decode(&deploymentsA)
	if deploymentsA[0].Sha != revA.Sha || len(deploymentsA) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsA)
	}

	request, _ = http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", projectB.ApiToken)
	response = httptest.NewRecorder()

	base.ListDeployments(response, request, projectB)

	decoder = json.NewDecoder(response.Body)

	var deploymentsB []Deployments
	_ = decoder.Decode(&deploymentsB)
	if deploymentsB[0].Sha != revB.Sha || len(deploymentsB) > 1 {
		t.Fatalf("Received foreign deployment: %v", deploymentsB)
	}
}

func TestListDeploymentsReturnsValidJSON(t *testing.T) {
	project := CreateTestProject("")

	var deploy Deployments = CreateTestDeployment(project, "test")

	request, _ := http.NewRequest("GET", "/deployments", nil)
	request.Header.Set("API-TOKEN", project.ApiToken)
	response := httptest.NewRecorder()

	base.ListDeployments(response, request, project)

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
