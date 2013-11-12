package main

import (
	"database/sql"
	_ "github.com/lib/pq"

	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Deploy struct {
	DeployedAt time.Time
	Sha        string
}

func ListRevisions(w http.ResponseWriter, req *http.Request) {
	// var revisions []Deploy = make([]Deploy, 0)
 //  _, err := dbmap.Select(&revisions, "SELECT * FROM revisioneer.deployments ORDER BY deployed_at")
 //  if err != nil {
 //  	log.Fatal(err)
 //  }
	// b, err := json.Marshal(revisions)
	// w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// if err == nil {
	// 	io.WriteString(w, string(b))
	// } else {
	// 	io.WriteString(w, "[]")
	// }
	io.WriteString(w, "[]")
}

func CreateRevision(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)

	var deploy Deploy
	err := dec.Decode(&deploy)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	} else {
		deploy.DeployedAt = time.Now()
	}

	// revisions = append(revisions, deploy)
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
	db, err := sql.Open("postgres", "user=nicolai86 dbname=revisioneer sslmode=disable")
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM revisioneer.deployments`)
  if err != nil {
    fmt.Println("Failed to run query", err)
    return
  }
  cols, err := rows.Columns()
  if err != nil {
    fmt.Println("Failed to get columns", err)
    return
  }

  // Result is your slice string.
  rawResult := make([][]byte, len(cols))

  dest := make([]interface{}, len(cols)) // A temporary interface{} slice
  for i, _ := range rawResult {
      dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
  }

  for rows.Next() {
  	var id int64
  	var sha string
  	var date time.Time
    err = rows.Scan(&id, &sha, &date)
    if err != nil {
        fmt.Println("Failed to scan row", err)
        return
    }

    fmt.Printf("%#v%#v%#v\n", id, sha, date)
  }

	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
