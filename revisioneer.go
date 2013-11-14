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

func (m *Messages) UnmarshalJSON(data []byte) error {
	if m == nil {
		*m = Messages{}
	}

	var message string
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	(*m).Message = message

	return nil
}

type Deployments struct {
	Id         hood.Id    `json:"-"`
	Sha        string     `json:"sha"`
	DeployedAt time.Time  `json:"deployed_at"`
	ProjectId  int        `json:"-"`
	Messages   []Messages `sql:"-" json:"messages"`
}

type Projects struct {
	Id        hood.Id   `json:"-"`
	Name      string    `json:"name"`
	ApiToken  string    `json:"api_token"`
	CreatedAt time.Time `json:"created_at"`
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
	if err := hd.Where("project_id", "=", project.Id).OrderBy("deployed_at").Find(&deployments); err != nil {
		log.Fatal("unable to load deployments", err)
	}

	// load messages for each deployment. N+1 queries
	for i, deployment := range deployments {
		hd.Where("deployment_id", "=", deployment.Id).Find(&deployments[i].Messages)
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

func CreateDeployment(w http.ResponseWriter, req *http.Request) {
	hd := Hd()
	defer hd.Db.Close()

	project, error := RequireProject(req)
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

	_, err := hd.Save(&deploy)
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
	bytes := make([]byte, STRLEN)
	rand.Read(bytes)

	encoding := base64.StdEncoding
	encoded := make([]byte, encoding.EncodedLen(len(bytes)))
	encoding.Encode(encoded, bytes)

	return string(encoded)
}

func CreateProject(w http.ResponseWriter, req *http.Request) {
	hd := Hd()
	defer hd.Db.Close()

	dec := json.NewDecoder(req.Body)

	var project Projects
	if err := dec.Decode(&project); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		project.CreatedAt = time.Now()
	}
	project.ApiToken = GenerateApiToken()

	_, err := hd.Save(&project)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.Marshal(project)
	io.WriteString(w, string(b))
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/deployments", ListDeployments).
		Methods("GET")
	r.HandleFunc("/deployments", CreateDeployment).
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
