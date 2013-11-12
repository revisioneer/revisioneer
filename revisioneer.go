package main

import (
	_ "github.com/lib/pq"
	"github.com/eaigner/hood"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Deployments struct {
	Id hood.Id
	Sha        string
	DeployedAt time.Time
}

var hd *hood.Hood
func Hd() (*hood.Hood) {
	if hd != nil {
		return hd
	}

	var err error
	hd, err = hood.Open("postgres", "user=nicolai86 dbname=revisioneer sslmode=disable")
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
		log.Fatal(err)
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
