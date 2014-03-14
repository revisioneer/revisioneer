package controllers

import (
	"database/sql"
	"github.com/eaigner/hood"
	_ "github.com/lib/pq"
	"log"
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
