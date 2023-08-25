package db

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Journey struct {
	bun.BaseModel `bun:"table:railmiles_journeys" json:"-"`

	ID uuid.UUID `bun:",pk,type:uuid" json:"id"`

	From     *StationName   `json:"from"`
	To       *StationName   `json:"to"`
	Via      []*StationName `bun:",nullzero" json:"via"`
	Distance float32        `json:"distance"`
	Date     time.Time      `json:"date"`
	Return   bool           `json:"return"`
}

type Route struct {
	bun.BaseModel `bun:"table:railmiles_routes"`

	From  string
	To    string
	Route []string `bun:",nullzero"`
}

type StationName struct {
	Shortcode string
	Full      string
}

func (sn *StationName) UnmarshalJSON(x []byte) error {
	if bytes.Equal(x, []byte("null")) {
		return nil
	}
	return json.Unmarshal(x, &sn.Shortcode)
}

func (sn *StationName) MarshalJSON() ([]byte, error) {
	if sn.Full == "" {
		return json.Marshal(sn.Shortcode)
	} else {
		return json.Marshal(map[string]string{"full": sn.Full, "shortcode": sn.Shortcode})
	}
}

func (sn *StationName) Scan(src any) error {
	if t, ok := src.(string); !ok {
		return fmt.Errorf("(*StationName).Scan can only read strings (not %T)", src)
	} else {
		sn.Shortcode = t
	}
	return nil
}

func (sn *StationName) Value() (driver.Value, error) {
	return sn.Shortcode, nil
}
