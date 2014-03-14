package main

import (
	. "./app/controllers"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
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

	defer base.Hd.Db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/deployments", base.WithValidProject(base.ListDeployments)).
		Methods("GET")
	r.HandleFunc("/deployments", base.WithValidProject(base.CreateDeployment)).
		Methods("POST")
	r.HandleFunc("/projects", base.CreateProject).
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
