package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var server *httptest.Server
var p project
var d deployment

func init() {
	server = httptest.NewServer(NewServer())

	DB.Exec(`truncate table projects, messages, deployments;`)
}

func TestCreateProject(t *testing.T) {
	resp, err := http.Post(server.URL+"/projects", "application/json", strings.NewReader(`{"name": "example"}`))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %v", resp.Body)
	}

	dec := json.NewDecoder(resp.Body)
	dec.Decode(&p)

	if p.Name != "example" {
		t.Fatalf("Created project /w wrong name")
	}

	if p.APIToken == "" {
		t.Fatalf("Created project wo/ wrong api token")
	}
}

func TestCreateDeployment(t *testing.T) {
	req, _ := http.NewRequest("POST", server.URL+"/deployments", strings.NewReader(`{
		"sha": "61722b0020",
		"messages": [
			"+ added support for messages",
			"Initial Commit"
		],
		"new_commit_counter": 2
	}`))
	req.Header.Set("API-Token", p.APIToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create deployment: %v, %d", resp.Body, resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	dec.Decode(&d)

	if d.NewCommitCounter != 2 {
		t.Fatalf("Wrong new commit counter: %d", d.NewCommitCounter)
	}

	if d.Sha != "61722b0020" {
		t.Fatalf("Wrong commit SHA: %v", d.Sha)
	}

	if d.Verified {
		t.Fatalf("Not verified yet")
	}
}

func TestVerifyDeployment(t *testing.T) {
	req, _ := http.NewRequest("POST", server.URL+"/deployments/"+d.Sha+"/verify", nil)
	req.Header.Set("API-Token", p.APIToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to verify deployment")
	}

	dec := json.NewDecoder(resp.Body)
	dec.Decode(&d)

	if !d.Verified {
		t.Fatalf("Should have been verified")
	}
}

func TestListDeployments(t *testing.T) {
	req, _ := http.NewRequest("GET", server.URL+"/deployments", nil)
	req.Header.Set("API-Token", p.APIToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unable to list deployments")
	}

	var ds []deployment
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&ds)

	if len(ds) != 1 {
		t.Fatalf("Wrong number of deployments")
	}
}
