package httpsrv

import (
	"encoding/json"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/gofiber/fiber/v2"
)

func (hs *httpServer) dashboardInfo(ctx *fiber.Ctx) error {
	var response = struct {
		GeoJSON json.RawMessage `json:"geoJSON,omitempty"`
		Stats   struct {
			LastMonth *core.JourneyStats `json:"lastMonth"`
			YTD       *core.JourneyStats `json:"ytd"`
			AllTime   *core.JourneyStats `json:"allTime"`
		} `json:"stats"`
		Journeys []*db.Journey `json:"journeys"`
	}{}

	journeys, err := hs.core.GetJourneys(&core.GetJourneysArgs{Since: core.LastMonth})
	if err != nil {
		return util.Wrap(err, "fetching journeys in the last month")
	}

	response.GeoJSON = []byte(hs.core.GenerateJourneyGeoJSON(journeys))

	lastMonthStats, err := hs.core.GetJourneyStats(core.LastMonth)
	if err != nil {
		return util.Wrap(err, "fetching last month stats")
	}

	ytdStats, err := hs.core.GetJourneyStats(core.YearToDate)
	if err != nil {
		return util.Wrap(err, "fetching year-to-date stats")
	}

	allTimeStats, err := hs.core.GetJourneyStats(core.AllTime)
	if err != nil {
		return util.Wrap(err, "fetching all time stats")
	}

	response.Stats.LastMonth = lastMonthStats
	response.Stats.YTD = ytdStats
	response.Stats.AllTime = allTimeStats

	core.PopulateFullStationNames(journeys)
	response.Journeys = journeys

	return ctx.JSON(response)
}
