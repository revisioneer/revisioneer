package controllers

import (
	. "../models"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/eaigner/hood"
	"io"
	"log"
	"net/http"
	"time"
)

const STRLEN = 32

var hd *hood.Hood

func GenerateApiToken() string {
	bytes := make([]byte, STRLEN)
	rand.Read(bytes)

	encoding := base64.StdEncoding
	encoded := make([]byte, encoding.EncodedLen(len(bytes)))
	encoding.Encode(encoded, bytes)

	return string(encoded)
}

func CreateProject(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)

	var project Projects
	if err := dec.Decode(&project); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		project.CreatedAt = time.Now()
	}
	project.ApiToken = GenerateApiToken()

	_, err := hd.Save(&project)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.Marshal(project)
	io.WriteString(w, string(b))
}
