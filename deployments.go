package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/eaigner/jet"
	"github.com/gorilla/mux"
)

type Deployment struct {
	Id               int       `json:"-"`
	Sha              string    `json:"sha"`
	DeployedAt       time.Time `json:"deployed_at"`
	ProjectId        int       `json:"-"`
	NewCommitCounter int       `json:"new_commit_counter"`
	Messages         []Message `sql:"-" json:"messages"`
	Verified         bool      `json:"verified"`
	VerifiedAt       time.Time `json:"verified_at"`
}

func (d *Deployment) Store(db *jet.Db) bool {
	var err error
	if d.Id != 0 {
		err = db.Query(`UPDATE deployments SET
			sha = $1,
			deployed_at = $2,
			new_commit_counter = $3,
			verified = $4, verified_at = $5
	 WHERE id = $6`,
			d.Sha,
			d.DeployedAt,
			d.NewCommitCounter,
			d.Verified,
			d.VerifiedAt,
			d.Id).Run()
	} else {
		err = db.Query(`INSERT INTO
			deployments
			(sha, deployed_at, project_id, new_commit_counter, verified, verified_at)
			VALUES
			($1, $2, $3, $4, $5, $6) RETURNING *`,
			d.Sha, d.DeployedAt,
			d.ProjectId, d.NewCommitCounter,
			d.Verified, d.VerifiedAt).Rows(d)
	}
	if err != nil {
		fmt.Printf(`%v: %#v`, err, d)
	}
	return err == nil
}

type DeploymentsController struct {
	*jet.Db
}

func (base *DeploymentsController) WithValidProject(next func(http.ResponseWriter, *http.Request, Project)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apiToken := req.Header.Get("API-TOKEN")

		var project Project
		base.Query(`SELECT * FROM projects WHERE api_token = $1 LIMIT 1`, apiToken).Rows(&project)

		if project == (Project{}) {
			http.Error(w, "unknown api token/ project", 500)
			return
		}

		next(w, req, project)
	}
}

func (base *DeploymentsController) WithValidProjectAndParams(next func(http.ResponseWriter, *http.Request, Project, map[string]string)) func(http.ResponseWriter, *http.Request) {
	return base.WithValidProject(func(w http.ResponseWriter, req *http.Request, project Project) {
		vars := mux.Vars(req)
		next(w, req, project, vars)
	})
}

func NewDeploymentsController(base *jet.Db) *DeploymentsController {
	return &DeploymentsController{Db: base}
}

func (controller *DeploymentsController) ListDeployments(w http.ResponseWriter, req *http.Request, project Project) {
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
	var deployments []Deployment
	if err = controller.
		Query(`SELECT * FROM deployments
			WHERE project_id = $1
			ORDER BY deployed_at DESC
			OFFSET $2 LIMIT $3`, project.Id, (page-1)*limit, limit).
		Rows(&deployments); err != nil {
		log.Fatal("unable to load deployments", err)
	}

	// load messages for each deployment. N+1 queries
	for i, deployment := range deployments {
		controller.Query(`SELECT * FROM messages WHERE deployment_id = $1`, deployment.Id).Rows(&deployments[i].Messages)
		if len(deployments[i].Messages) == 0 {
			deployments[i].Messages = make([]Message, 0)
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

func (controller *DeploymentsController) VerifyDeployment(w http.ResponseWriter, req *http.Request, project Project, vars map[string]string) {
	var deployment Deployment
	controller.Query(`SELECT * FROM deployments WHERE sha = $1 LIMIT 1`, vars["sha"]).Rows(&deployment)

	if reflect.DeepEqual(deployment, Deployment{}) {
		http.Error(w, "unknown deployment revision", 404)
		return
	}

	if !deployment.Verified {
		deployment.Verified = true
		deployment.VerifiedAt = time.Now()

		if !deployment.Store(controller.Db) {
			log.Fatalf(`unable to mark deployment as verified`)
		}
	}

	b, err := json.Marshal(deployment)

	if err == nil {
		io.WriteString(w, string(b))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{}")
	}
}

func (controller *DeploymentsController) CreateDeployment(w http.ResponseWriter, req *http.Request, project Project) {
	dec := json.NewDecoder(req.Body)

	var deploy Deployment
	if err := dec.Decode(&deploy); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}
	deploy.Verified = false

	deploy.ProjectId = project.Id

	// TODO wrap in transaction
	if !deploy.Store(controller.Db) {
		log.Fatal("Unable to create deployment")
	}

	for _, message := range deploy.Messages {
		message.DeploymentId = int(deploy.Id)
		if !message.Store(controller.Db) {
			log.Fatal("Unable to save message")
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
