package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

type message struct {
	ID           int `json:"-"`
	Message      string
	DeploymentID int `json:"-"`
}

func (m message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Message)
}

func (m *message) UnmarshalJSON(data []byte) error {
	if m == nil {
		m = &message{}
	}

	if err := json.Unmarshal(data, &m.Message); err != nil {
		return err
	}

	return nil
}

func (m *message) Store(db *sql.DB) bool {
	var err error
	if m.ID != 0 {
		_, err = db.Exec(`UPDATE messages SET message = $1 WHERE id = $2`, m.Message, m.ID)
	} else {
		err = db.QueryRow(`INSERT INTO messages (message, deployment_id) VALUES ($1, $2) RETURNING id`, m.Message, m.DeploymentID).Scan(&m.ID)
	}
	return err == nil
}

type project struct {
	ID        int       `json:"-"`
	Name      string    `json:"name"`
	APIToken  string    `json:"api_token"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *project) Store(db *sql.DB) bool {
	var err error
	if p.ID != 0 {
		_, err = db.Exec(`UPDATE projects SET WHERE id = $1`, p.ID)
	} else {
		err = db.QueryRow(`INSERT INTO projects
			(name, api_token, created_at)
			VALUES
			($1, $2, NOW()) RETURNING id`, p.Name, p.APIToken).Scan(&p.ID)
	}
	return err == nil
}

func (p *project) IsValid(db *sql.DB) bool {
	invalidData := p.Name == "" || p.APIToken == ""

	exists := new(bool)
	db.QueryRow(`select 't' from projects where api_token = '$1' limit 1`, p.APIToken).Scan(&exists)

	return !(*exists || invalidData)
}

type projectsController struct {
	*sql.DB
}

func newProjectsController(base *sql.DB) *projectsController {
	return &projectsController{DB: base}
}

// STRLEN defines how long the generated APITokens are
const STRLEN = 32

func generateAPIToken() string {
	bytes := make([]byte, STRLEN)
	rand.Read(bytes)

	encoding := base64.StdEncoding
	encoded := make([]byte, encoding.EncodedLen(len(bytes)))
	encoding.Encode(encoded, bytes)

	return string(encoded)
}

func (controller *projectsController) createProject(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)

	var p project
	if err := dec.Decode(&p); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		p.CreatedAt = time.Now()
	}
	p.APIToken = generateAPIToken()

	for i := 0; i < 10; i++ {
		if !p.IsValid(controller.DB) {
			p.APIToken = generateAPIToken()
		} else {
			break
		}
	}

	if !p.IsValid(controller.DB) {
		log.Fatal("project is not valid. %v", p)
	}

	if !p.Store(controller.DB) {
		log.Fatal("unable to create project")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(p)
}

type deployment struct {
	ID               int       `json:"-"`
	Sha              string    `json:"sha"`
	DeployedAt       time.Time `json:"deployed_at"`
	ProjectID        int       `json:"-"`
	NewCommitCounter int       `json:"new_commit_counter"`
	Messages         []message `json:"messages, omitempty"`
	Verified         bool      `json:"verified"`
	VerifiedAt       time.Time `json:"verified_at, omitempty"`
}

func (d *deployment) LoadMessages(db *sql.DB) error {
	var messages []message
	rows, err := db.Query(`SELECT id, message, deployment_id FROM messages WHERE deployment_id = $1`, d.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var m message
		if rows.Scan(&m.ID, &m.Message, &m.DeploymentID); err != nil {
			return err
		}
		messages = append(messages, m)
	}
	d.Messages = messages
	return nil
}

func (d *deployment) Store(db *sql.DB) bool {
	var err error
	err = db.QueryRow(`INSERT INTO
			deployments
			(sha, deployed_at, project_id, new_commit_counter, verified, verified_at)
			VALUES
			($1, $2, $3, $4, $5, $6) RETURNING id, sha, deployed_at, project_id, new_commit_counter, verified, verified_at`,
		d.Sha, d.DeployedAt,
		d.ProjectID, d.NewCommitCounter,
		d.Verified, d.VerifiedAt).Scan(&d.ID, &d.Sha, &d.DeployedAt,
		&d.ProjectID, &d.NewCommitCounter,
		&d.Verified, &d.VerifiedAt)
	if err != nil {
		fmt.Printf(`%v: %#v`, err, d)
	}
	return err == nil
}

type deploymentsController struct {
	*sql.DB
}

func (base *deploymentsController) WithValidProject(next func(http.ResponseWriter, *http.Request, project)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apiToken := req.Header.Get("API-TOKEN")

		var p project
		if err := base.QueryRow(`SELECT
				id,
				name,
				created_at
			FROM projects
			WHERE api_token = $1`, apiToken).Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			http.Error(w, "unknown project", 500)
			return
		}

		if p == (project{}) {
			http.Error(w, "unknown api token/ project", 404)
			return
		}

		next(w, req, p)
	}
}

func (base *deploymentsController) WithValidProjectAndParams(next func(http.ResponseWriter, *http.Request, project, map[string]string)) func(http.ResponseWriter, *http.Request) {
	return base.WithValidProject(func(w http.ResponseWriter, req *http.Request, p project) {
		vars := mux.Vars(req)
		next(w, req, p, vars)
	})
}

func newDeploymentsController(base *sql.DB) *deploymentsController {
	return &deploymentsController{DB: base}
}

func (base *deploymentsController) listDeployments(w http.ResponseWriter, req *http.Request, p project) {
	limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
	if err != nil {
		limit = 20
	}
	limit = int(math.Min(math.Abs(float64(limit)), 100.0))

	var page int
	page, err = strconv.Atoi(req.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	page = int(math.Max(float64(page), 1.0))

	// load deployments
	var deployments []deployment
	rows, err := base.Query(`SELECT id, sha, deployed_at, project_id, new_commit_counter, verified, verified_at
			FROM deployments
			WHERE project_id = $1
			ORDER BY deployed_at DESC
			OFFSET $2 LIMIT $3`, p.ID, (page-1)*limit, limit)
	if err != nil {
		log.Fatal("unable to load deployments: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer rows.Close()

	// load deployments and messages for each deployment. N+1 queries
	for rows.Next() {
		d := deployment{}
		if err := rows.Scan(&d.ID, &d.Sha, &d.DeployedAt, &d.ProjectID, &d.NewCommitCounter, &d.Verified, &d.VerifiedAt); err != nil {
			log.Println("database error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		d.LoadMessages(base.DB)
		deployments = append(deployments, d)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(deployments); err != nil {
		log.Fatalf("unable to serialize deployments")
	}
}

func (base *deploymentsController) verifyDeployment(w http.ResponseWriter, req *http.Request, p project, vars map[string]string) {
	var d deployment
	base.QueryRow(`SELECT id FROM deployments WHERE sha = $1 LIMIT 1`, vars["sha"]).Scan(&d.ID)

	if reflect.DeepEqual(d, deployment{}) {
		http.Error(w, "unknown deployment revision", 404)
		return
	}

	if !d.Verified {
		d.Verified = true
		d.VerifiedAt = time.Now()

		if _, err := base.DB.Exec(`UPDATE deployments SET verified = 't', verified_at = NOW() WHERE id = $1`, d.ID); err != nil {
			log.Fatalf(`unable to mark deployment as verified: %v`, err)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	encoder := json.NewEncoder(w)
	encoder.Encode(d)
}

func (base *deploymentsController) createDeployment(w http.ResponseWriter, req *http.Request, p project) {
	dec := json.NewDecoder(req.Body)

	var deploy deployment
	if err := dec.Decode(&deploy); err != nil && err != io.EOF {
		log.Fatal("decode error", err)
	} else {
		deploy.DeployedAt = time.Now()
	}
	deploy.Verified = false
	deploy.ProjectID = p.ID

	if !deploy.Store(base.DB) {
		log.Fatal("Unable to create deployment")
	}

	for _, message := range deploy.Messages {
		message.DeploymentID = deploy.ID
		if !message.Store(base.DB) {
			log.Fatal("Unable to save message")
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(deploy)
}

func setup() *sql.DB {
	var dbName = os.Getenv("DATABASE")
	if dbName == "" {
		dbName = "revisioneer"
	}

	var revDsn = os.Getenv("REV_DSN")
	if revDsn == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		revDsn = "user=" + user.Username + " dbname=" + dbName + " sslmode=disable"
	}

	db, err := sql.Open("postgres", revDsn)
	if err != nil {
		log.Fatal("failed to connect to postgres", err)
	}
	db.SetMaxIdleConns(100)

	return db
}

func runMigrations(db *sql.DB) {
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "migrations",
	}

	if n, err := migrate.Exec(db, "postgres", migrations, migrate.Up); err != nil {
		log.Printf("unable to migrate: %v", err)
	} else {
		log.Printf("Applied %d migrations!\n", n)
	}
}

// Set by make file on build
var (
	Version string
	Commit  string
	DB      *sql.DB
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))
	DB = setup()
}

func logHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}

func NewServer() http.Handler {
	runMigrations(DB)

	deployments := newDeploymentsController(DB)
	projects := newProjectsController(DB)

	r := mux.NewRouter()
	r.HandleFunc("/deployments", deployments.WithValidProject(deployments.listDeployments)).
		Methods("GET")
	r.HandleFunc("/deployments", deployments.WithValidProject(deployments.createDeployment)).
		Methods("POST")
	r.HandleFunc("/deployments/{sha}/verify", deployments.WithValidProjectAndParams(deployments.verifyDeployment)).
		Methods("POST")
	r.HandleFunc("/projects", projects.createProject).
		Methods("POST")
	return http.Handler(logHandler(r))
}

func main() {
	var (
		httpAddress  = flag.String("http.addr", ":8080", "HTTP listen address")
		printVersion = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Printf("%s", Version)
		os.Exit(0)
	}

	log.Printf("listening on %s", *httpAddress)

	http.ListenAndServe(*httpAddress, NewServer())
}
