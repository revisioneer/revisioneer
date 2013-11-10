package main

import (
  "io"
  "time"
  "github.com/bmizerany/pat"
  "net/http"
)

type Deploy struct {
  DeployedAt time.Time
  Sha string
}

func ListRevisions(w http.ResponseWriter, req *http.Request) {
  io.WriteString(w, "revisions: \n")
}

func CreateRevision(w http.ResponseWriter, req *http.Request) {
  io.WriteString(w, "created!\n")
}

func init() {
  muxer := pat.New()
  muxer.Get("/revisions", http.HandlerFunc(ListRevisions))
  muxer.Post("/revisions", http.HandlerFunc(CreateRevision))
  http.Handle("/", muxer)
}

func main() {
  http.ListenAndServe(":8080", nil)
}