package main

import (
	"encoding/json"
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

type Deployments struct {
	Id         hood.Id
	Sha        string
	DeployedAt time.Time
}

var hd *hood.Hood

func Hd() *hood.Hood {
	if hd != nil {
		return hd
	}

	var revDsn = os.Getenv("REV_DSN")
	if revDsn == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		revDsn = "user=" + user.Username + " dbname=revisioneer sslmode=disable"
	}

	var err error
	hd, err = hood.Open("postgres", revDsn)
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	return hd
}

func ListRevisions(w http.ResponseWriter, req *http.Request) {
	var revisions []Deployments
	err := Hd().OrderBy("deployed_at").Find(&revisions)
	if err != nil {
		log.Fatal("unable to load deployments", err)
	}

	b, err := json.Marshal(revisions)
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
	dec := json.NewDecoder(req.Body)

	var deploy Deployments
	err := dec.Decode(&deploy)
	if err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}

	_, err = Hd().Save(&deploy)
	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, "")
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/revisions", ListRevisions).
		Methods("GET")
	r.HandleFunc("/revisions", CreateRevision).
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
