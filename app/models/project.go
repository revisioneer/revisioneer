package models

import (
	"github.com/eaigner/hood"
	"time"
)

type Projects struct {
	Id        hood.Id   `json:"-"`
	Name      string    `json:"name"`
	ApiToken  string    `json:"api_token"`
	CreatedAt time.Time `json:"created_at"`
}
