package core

import (
	_ "embed"
	"encoding/json"
	"github.com/codemicro/railmiles/railmiles/internal/config"
	"github.com/codemicro/railmiles/railmiles/internal/db"
)

type Core struct {
	config *config.Config
	db     *db.DB
}

func New(conf *config.Config, database *db.DB) *Core {
	return &Core{
		config: conf,
		db:     database,
	}
}

//go:embed stationData.json
var stationDataRaw []byte

type StationDetail struct {
	Name string
	Lat  float32
	Lon  float32
}

var (
	stationData map[string]*StationDetail
)

func init() {
	_ = json.Unmarshal(stationDataRaw, &stationData)
}

func GetStationName(short string) string {
	x := short
	if ff, found := stationData[short]; found {
		x = ff.Name
	}
	return x
}

func GetStationDetail(short string) *StationDetail {
	if x, found := stationData[short]; found {
		return x
	}
	return nil
}
