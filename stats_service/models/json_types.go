package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type FloatSlice []float64

func (fs *FloatSlice) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Scan failed: %v", value)
	}
	return json.Unmarshal(b, &fs)
}

func (fs FloatSlice) Value() (driver.Value, error) {
	return json.Marshal(fs)
}

type JSONMarkScale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

func (j *JSONMarkScale) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Scan failed: %v", value)
	}
	return json.Unmarshal(b, &j)
}

func (j JSONMarkScale) Value() (driver.Value, error) {
	return json.Marshal(j)
}
