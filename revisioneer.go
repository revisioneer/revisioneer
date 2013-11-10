package main

import (
  "fmt"
  "time"
  "net/http"
)

type Deploy struct {
  DeployedAt time.Time
  Sha string
}

func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}