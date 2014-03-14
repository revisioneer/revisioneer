package models

import (
	"encoding/json"
	"github.com/eaigner/hood"
)

type Messages struct {
	Id           hood.Id
	Message      string
	DeploymentId int
}

func (m Messages) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Message)
}

func (m *Messages) UnmarshalJSON(data []byte) error {
	if m == nil {
		*m = Messages{}
	}

	var message string
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	(*m).Message = message

	return nil
}
