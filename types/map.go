package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Map map[string]interface{}

func (m Map) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Map) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &m)
}
