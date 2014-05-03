package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	. "github.com/revisioneer/revisioneer/controllers"
)

var base *Base

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))

	base = &Base{}
	base.Setup()
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

	defer base.Db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/deployments", base.WithValidProject(NewDeploymentsController(base).ListDeployments)).
		Methods("GET")
	r.HandleFunc("/deployments", base.WithValidProject(NewDeploymentsController(base).CreateDeployment)).
		Methods("POST")
	r.HandleFunc("/deployments/{sha}/verify", base.WithValidProjectAndParams(NewDeploymentsController(base).VerifyDeployment)).
		Methods("POST")
	r.HandleFunc("/projects", NewProjectsController(base).CreateProject).
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
