package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eaigner/hood"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"
)

type Messages struct {
	Id           hood.Id
	Message      string
	DeploymentId int
}

func (m Messages) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Message)
}

type Deployments struct {
	Id         hood.Id    `json:"-"`
	Sha        string     `json:"sha"`
	DeployedAt time.Time  `json:"deployed_at"`
	ProjectId  int        `json:"-"`
	Messages   []Messages `sql:"-" json:"messages"`
}

type Projects struct {
	Id        hood.Id
	Name      string
	ApiToken  string
	CreatedAt time.Time
}

func Hd() *hood.Hood {
	var revDsn = os.Getenv("REV_DSN")
	if revDsn == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		revDsn = "user=" + user.Username + " dbname=revisioneer sslmode=disable"
	}

	var err error
	hd, err := hood.Open("postgres", revDsn)
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	hd.Log = true
	return hd
}

func RequireProject(req *http.Request) (Projects, error) {
	hd := Hd()
	defer hd.Db.Close()

	apiToken := req.Header.Get("API-TOKEN")
	var projects []Projects
	hd.Where("api_token", "=", apiToken).Limit(1).Find(&projects)

	if len(projects) != 1 {
		return Projects{}, errors.New("Unknown project")
	}

	return projects[0], nil
}

func ListDeployments(w http.ResponseWriter, req *http.Request) {
	hd := Hd()
	defer hd.Db.Close()

	project, error := RequireProject(req)
	if error != nil {
		http.Error(w, "unknown api token/ project", 500)
		return
	}

	// load deployments
	var deployments []Deployments
	err := hd.Where("project_id", "=", project.Id).OrderBy("deployed_at").Find(&deployments)
	if err != nil {
		log.Fatal("unable to load deployments", err)
	}

	// load messages for each deployment. N+1 queries
	for i, revision := range deployments {
		hd.Where("deployment_id", "=", revision.Id).Find(&deployments[i].Messages)
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

func CreateRevision(w http.ResponseWriter, req *http.Request) {
	hd := Hd()
	defer hd.Db.Close()

	project, error := RequireProject(req)
	if error != nil {
		http.Error(w, "unknown api token/ project", 500)
		return
	}

	dec := json.NewDecoder(req.Body)

	var deploy Deployments
	err := dec.Decode(&deploy)
	if err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}
	deploy.ProjectId = int(project.Id)

	_, err = hd.Save(&deploy)
	if err != nil {
		log.Fatal(err)
	}

	for _, message := range deploy.Messages {
		message.DeploymentId = int(deploy.Id)
		_, err = hd.Save(&message)
		if err != nil {
			log.Fatal(err)
		}
	}

	io.WriteString(w, "")
}

const STRLEN = 32

func GenerateApiToken() string {
	b := make([]byte, STRLEN)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}

func CreateProject(w http.ResponseWriter, req *http.Request) {
	hd := Hd()
	defer hd.Db.Close()

	dec := json.NewDecoder(req.Body)

	var project Projects
	err := dec.Decode(&project)
	if err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		project.CreatedAt = time.Now()
	}
	project.ApiToken = GenerateApiToken()

	_, err = hd.Save(&project)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.Marshal(project)
	io.WriteString(w, string(b))
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/revisions", ListDeployments).
		Methods("GET")
	r.HandleFunc("/revisions", CreateRevision).
		Methods("POST")
	r.HandleFunc("/projects", CreateProject).
		Methods("POST")
	http.Handle("/", r)
}

func main() {
	Hd()

	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}
