package controllers

import (
	. "../models"
	"encoding/json"
	"errors"
	_ "github.com/eaigner/hood"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

func (base *Base) RequireProject(req *http.Request) (Projects, error) {
	apiToken := req.Header.Get("API-TOKEN")
	var projects []Projects
	base.Hd.Where("api_token", "=", apiToken).Limit(1).Find(&projects)

	if len(projects) != 1 {
		return Projects{}, errors.New("Unknown project")
	}

	return projects[0], nil
}

func (base *Base) ListDeployments(w http.ResponseWriter, req *http.Request) {
	project, error := base.RequireProject(req)
	if error != nil {
		http.Error(w, "unknown api token/ project", 500)
		return
	}

	limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
	if err != nil {
		limit = 20
	}
	limit = int(math.Min(math.Abs(float64(limit)), 100.0))

	var page int
	page, err = strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	page = int(math.Max(float64(page), 1.0))

	// load deployments
	var deployments []Deployments
	err = base.Hd.
		Where("project_id", "=", project.Id).
		OrderBy("deployed_at").
		Desc().
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&deployments)
	if err != nil {
		log.Fatal("unable to load deployments", err)
	}

	// load messages for each deployment. N+1 queries
	for i, deployment := range deployments {
		base.Hd.Where("deployment_id", "=", deployment.Id).Find(&deployments[i].Messages)
		if len(deployments[i].Messages) == 0 {
			deployments[i].Messages = make([]Messages, 0)
		}
	}

	b, err := json.Marshal(deployments)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err == nil {
		if string(b) == "null" {
			io.WriteString(w, "[]")
		} else {
			io.WriteString(w, string(b))
		}

	} else {
		io.WriteString(w, "[]")
	}
}

func (base *Base) CreateDeployment(w http.ResponseWriter, req *http.Request) {
	project, error := base.RequireProject(req)
	if error != nil {
		http.Error(w, "unknown api token/ project", 500)
		return
	}

	dec := json.NewDecoder(req.Body)

	var deploy Deployments
	if err := dec.Decode(&deploy); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}
	deploy.ProjectId = int(project.Id)

	_, err := base.Hd.Save(&deploy)
	if err != nil {
		log.Fatal(err)
	}

	for _, message := range deploy.Messages {
		message.DeploymentId = int(deploy.Id)
		_, err = base.Hd.Save(&message)
		if err != nil {
			log.Fatal(err)
		}
	}

	io.WriteString(w, "")
}