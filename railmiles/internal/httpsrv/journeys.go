package httpsrv

import (
	"encoding/json"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"math"
	"strconv"
)

func (hs *httpServer) journeyListing(ctx *fiber.Ctx) error {
	const pageSize = 20

	var pageNumber uint
	{
		pageNumberStr := ctx.Query("page", "0")
		pageNumber64, _ := strconv.ParseUint(pageNumberStr, 10, 32)
		pageNumber = uint(pageNumber64)
	}

	var response = struct {
		NumPages   int           `json:"numPages"`
		PageNumber uint          `json:"pageNumber"`
		Data       []*db.Journey `json:"data"`
	}{
		PageNumber: pageNumber,
	}

	journeyStats, err := hs.core.GetJourneyStats(core.AllTime)
	if err != nil {
		return util.Wrap(err, "getting all journey stats")
	}

	response.NumPages = int(math.Ceil(float64(journeyStats.RawCount/pageSize))) + 1

	if !(int(pageNumber*pageSize) > journeyStats.RawCount) {
		journeys, err := hs.core.GetJourneys(&core.GetJourneysArgs{Offset: int(pageSize * pageNumber), Limit: pageSize})
		if err != nil {
			return util.Wrap(err, "getting paginated journeys")
		}
		core.PopulateFullStationNames(journeys)
		response.Data = journeys
	}

	return ctx.JSON(response)
}

func (hs *httpServer) getJourney(ctx *fiber.Ctx) error {
	var response = struct {
		GeoJSON json.RawMessage `json:"geoJSON"`
		Data    *db.Journey     `json:"data"`
	}{}

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.ErrNotFound
	}

	journey, err := hs.core.GetJourney(id)
	if err != nil {
		return util.Wrap(err, "fetching journey %s", id.String())
	}

	if journey == nil {
		return fiber.ErrNotFound
	}

	ja := []*db.Journey{journey}
	core.PopulateFullStationNames(ja)

	response.Data = journey
	response.GeoJSON = []byte(hs.core.GenerateJourneyGeoJSON(ja, true))

	return ctx.JSON(&response)
}

func (hs *httpServer) deleteJourney(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return fiber.ErrNotFound
	}

	if err := hs.core.DeleteJourney(id); err != nil {
		return util.Wrap(err, "deleting journey %s", id.String())
	}

	ctx.Status(204)
	return nil
}
