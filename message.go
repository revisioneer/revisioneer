package main

import (
	"encoding/json"

	"github.com/splicers/jet"
)

type Message struct {
	Id           int
	Message      string
	DeploymentId int
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Message)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	if m == nil {
		m = &Message{}
	}

	if err := json.Unmarshal(data, &m.Message); err != nil {
		return err
	}

	return nil
}

func (m *Message) Store(db *jet.Db) bool {
	var err error
	if m.Id != 0 {
		err = db.Query(`UPDATE messages SET message = $1 WHERE id = $2`, m.Message, m.Id).Run()
	} else {
		err = db.Query(`INSERT INTO messages (message, deployment_id) VALUES ($1, $2) RETURNING *`, m.Message, m.DeploymentId).Rows(m)
	}
	return err == nil
}
