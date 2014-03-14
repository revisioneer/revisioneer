package main

import (
	. "./app/controllers"
	. "./app/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eaigner/hood"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

var db *sql.DB
var hd *hood.Hood

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
	db, err = sql.Open("postgres", revDsn)
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	db.SetMaxIdleConns(100)

	newHd := hood.New(db, hood.NewPostgres())
	newHd.Log = true
	return newHd
}

func RequireProject(req *http.Request) (Projects, error) {
	apiToken := req.Header.Get("API-TOKEN")
	var projects []Projects
	hd.Where("api_token", "=", apiToken).Limit(1).Find(&projects)

	if len(projects) != 1 {
		return Projects{}, errors.New("Unknown project")
	}

	return projects[0], nil
}

func ListDeployments(w http.ResponseWriter, req *http.Request) {
	project, error := RequireProject(req)
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
	err = hd.
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

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))

	hd = Hd()
}

func main() {
	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Listen on a TCP or a UNIX domain socket (TCP here).
	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if nil != err {
		log.Fatalln(err)
	}
	log.Printf("listening on %v", l.Addr())

	writePid()

	defer hd.Db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/deployments", ListDeployments).
		Methods("GET")
	r.HandleFunc("/deployments", CreateDeployment).
		Methods("POST")
	r.HandleFunc("/projects", CreateProject).
		Methods("POST")
	http.Handle("/", r)

	http.Serve(l, r)
}

func writePid() {
	var file, error = os.OpenFile("tmp/rev.pid", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if error == nil {
		var line = fmt.Sprintf("%v", os.Getpid())
		file.WriteString(line)
		file.Close()
	}
}
