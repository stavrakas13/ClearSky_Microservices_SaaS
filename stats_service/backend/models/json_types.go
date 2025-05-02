package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ✅ Υποστηρίζει JSONB για πίνακα βαθμών
type FloatSlice []float64

func (fs *FloatSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(bytes, fs)
}

func (fs FloatSlice) Value() (driver.Value, error) {
	return json.Marshal(fs)
}

// ✅ Υποστηρίζει JSONB για mark_scale { "min": ..., "max": ... }
type MarkScale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

func (ms *MarkScale) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", value)
	}
	return json.Unmarshal(bytes, ms)
}

func (ms MarkScale) Value() (driver.Value, error) {
	return json.Marshal(ms)
}
