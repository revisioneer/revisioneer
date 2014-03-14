package controllers

import (
	. "../models"
	"database/sql"
	"github.com/eaigner/hood"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/user"
)

type Base struct {
	Hd *hood.Hood
}

func (base *Base) Setup() {
	var revDsn = os.Getenv("REV_DSN")
	if revDsn == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		revDsn = "user=" + user.Username + " dbname=revisioneer sslmode=disable"
	}

	db, err := sql.Open("postgres", revDsn)
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	db.SetMaxIdleConns(100)

	newHd := hood.New(db, hood.NewPostgres())
	newHd.Log = true
	base.Hd = newHd
}

func (base *Base) WithValidProject(next func(http.ResponseWriter, *http.Request, Projects)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apiToken := req.Header.Get("API-TOKEN")
		var projects []Projects
		base.Hd.Where("api_token", "=", apiToken).Limit(1).Find(&projects)

		if len(projects) != 1 {
			http.Error(w, "unknown api token/ project", 500)
			return
		}

		next(w, req, projects[0])
	}
}
