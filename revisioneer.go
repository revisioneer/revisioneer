package main

import (
  "io"
  "time"
  "github.com/bmizerany/pat"
  "net/http"
  "encoding/json"
)

type Deploy struct {
  DeployedAt time.Time
  Sha string
}

var revisions []Deploy = make([]Deploy, 0)

func ListRevisions(w http.ResponseWriter, req *http.Request) {
  b, err := json.Marshal(revisions)
  if err == nil {
    io.WriteString(w, string(b))
  } else {
    io.WriteString(w, "[]")
  }
}

func CreateRevision(w http.ResponseWriter, req *http.Request) {
  var sha string = req.URL.Query().Get(":sha")
  var newDeploy Deploy = Deploy{time.Now(), sha}
  revisions = append(revisions, newDeploy);
  io.WriteString(w, "")
}

func init() {
  muxer := pat.New()
  muxer.Get("/revisions", http.HandlerFunc(ListRevisions))
  muxer.Post("/revisions/:sha", http.HandlerFunc(CreateRevision))
  http.Handle("/", muxer)
}

func main() {
  http.ListenAndServe(":8080", nil)
}