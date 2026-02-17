package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Coordinates represents geographic coordinates in decimal format.
// Latitude and Longitude are expressed in decimal degrees as per the
// World Geodetic System (WGS84).
type Coordinates struct {
	// Latitude is the latitude coordinate in decimal degrees (range: -90 to 90)
	Latitude float64 `json:"latitud"`

	// Longitude is the longitude coordinate in decimal degrees (range: -180 to 180)
	Longitude float64 `json:"longitud"`
}

// MeteocatTime parses time strings that may omit seconds (e.g., 1992-05-11T15:30Z).
// It marshals back to RFC3339 for stability in tests and consumers.
type MeteocatTime struct {
	time.Time
}

// UnmarshalJSON supports multiple METEOCAT time layouts, including values without seconds.
func (m *MeteocatTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04Z",
		"2006-01-02T15:04:05Z",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			m.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("parse time %q", raw)
}

// MarshalJSON writes the timestamp using RFC3339 format.
func (m MeteocatTime) MarshalJSON() ([]byte, error) {
	if m.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(m.Time.Format(time.RFC3339))
}
