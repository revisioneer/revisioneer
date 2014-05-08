package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	. "strings"
	"testing"

	_ "github.com/eaigner/hood"
)

var projectsController *ProjectsController

func init() {
	_db = Setup()
	projectsController = NewProjectsController(_db)
}
func ClearProjects() {
	_db.Query("DELETE FROM projects").Run()
}

func TestCreateProject(t *testing.T) {
	request, _ := http.NewRequest("POST", "/projects", NewReader(`{"name": "Musterprojekt"}`))
	response := httptest.NewRecorder()

	projectsController.CreateProject(response, request)

	decoder := json.NewDecoder(response.Body)

	var project Project
	err := decoder.Decode(&project)

	if err != nil {
		t.Fatalf("Decoding should pass: %v", err)
	}

	if project.Name != "Musterprojekt" {
		t.Fatalf("Name was set improperly. Expected %+v to %+v", "Musterprojekt", project.Name)
	}
	if project.ApiToken == "" {
		t.Fatalf("Expected ApiToken to be set to something")
	}
}
