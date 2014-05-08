package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/eaigner/jet"
)

type Project struct {
	Id        int       `json:"-"`
	Name      string    `json:"name"`
	ApiToken  string    `json:"api_token"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *Project) Store(db *jet.Db) bool {
	var err error
	if p.Id != 0 {
		err = db.Query(`UPDATE projects SET WHERE id = $1`, p.Id).Run()
	} else {
		err = db.Query(`INSERT INTO projects
			(name, api_token, created_at)
			VALUES
			($1, $2, NOW()) RETURNING *`, p.Name, p.ApiToken).Rows(p)
	}
	return err == nil
}

type ProjectsController struct {
	*jet.Db
}

func NewProjectsController(base *jet.Db) *ProjectsController {
	return &ProjectsController{Db: base}
}

const STRLEN = 32

func generateApiToken() string {
	bytes := make([]byte, STRLEN)
	rand.Read(bytes)

	encoding := base64.StdEncoding
	encoded := make([]byte, encoding.EncodedLen(len(bytes)))
	encoding.Encode(encoded, bytes)

	return string(encoded)
}

func (controller *ProjectsController) CreateProject(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)

	var project Project
	if err := dec.Decode(&project); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		project.CreatedAt = time.Now()
	}
	project.ApiToken = generateApiToken()

	// TODO loop until no collision on ApiToken exists

	if !project.Store(controller.Db) {
		log.Fatal("unable to create project")
	}

	b, _ := json.Marshal(project)
	io.WriteString(w, string(b))
}
