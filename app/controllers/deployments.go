package controllers

import (
	. "../models"
	"encoding/json"
	_ "github.com/eaigner/hood"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type DeploymentsController struct {
	Base *Base
}

func NewDeploymentsController(base *Base) *DeploymentsController {
	return &DeploymentsController{Base: base}
}

func (controller *DeploymentsController) ListDeployments(w http.ResponseWriter, req *http.Request, project Projects) {
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
	err = controller.Base.Hd.
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
		controller.Base.Hd.Where("deployment_id", "=", deployment.Id).Find(&deployments[i].Messages)
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

func (controller *DeploymentsController) VerifyDeployment(w http.ResponseWriter, req *http.Request, project Projects, vars map[string]string) {
	var deployments []Deployments
	controller.Base.Hd.Where("sha", "=", vars["sha"]).Find(&deployments)

	if len(deployments) != 1 {
		http.Error(w, "unknown deployment revision", 404)
		return
	}

	deployment := deployments[0]
	if !deployment.Verified {
		deployment.Verified = true
		deployment.VerifiedAt = time.Now()
		// deployment.VerifiedAt = Time(time.Now())

		controller.Base.Hd.Save(&deployment)
	}

	b, err := json.Marshal(deployment)

	if err == nil {
		io.WriteString(w, string(b))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{}")
	}
}

func (controller *DeploymentsController) CreateDeployment(w http.ResponseWriter, req *http.Request, project Projects) {
	dec := json.NewDecoder(req.Body)

	var deploy Deployments
	if err := dec.Decode(&deploy); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}
	deploy.Verified = false
	// deploy.VerifiedAt = Time(time.Time{})

	deploy.ProjectId = int(project.Id)

	_, err := controller.Base.Hd.Save(&deploy)
	if err != nil {
		log.Fatal(err)
	}

	for _, message := range deploy.Messages {
		message.DeploymentId = int(deploy.Id)
		_, err = controller.Base.Hd.Save(&message)
		if err != nil {
			log.Fatal(err)
		}
	}

	b, err := json.Marshal(deploy)

	if err == nil {
		io.WriteString(w, string(b))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{}")
	}
}
