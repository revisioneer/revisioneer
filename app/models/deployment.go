package models

import (
	"github.com/eaigner/hood"
	"time"
)

type Deployments struct {
	Id               hood.Id    `json:"-"`
	Sha              string     `json:"sha"`
	DeployedAt       time.Time  `json:"deployed_at"`
	ProjectId        int        `json:"-"`
	NewCommitCounter int        `json:"new_commit_counter"`
	Messages         []Messages `sql:"-" json:"messages"`
}
