package httpsrv

import (
	"encoding/json"
	"fmt"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"regexp"
	"strings"
	"time"
)

var locationsRegexp = regexp.MustCompile(`(?:(?:[A-Z]){3}(?:, ?[A-Z][0-9]{5})?\n?)+`)

func (hs *httpServer) newJourney(ctx *fiber.Ctx) error {
	var requestBody = struct {
		Date           time.Time `json:"date"`
		Route          string    `json:"route"`
		ManualDistance float32   `json:"manualDistance"`
		IsReturn       bool      `json:"isReturn"`
	}{}

	var response = struct {
		ID uuid.UUID `json:"id"`
	}{}

	if !strings.EqualFold(ctx.Get("Content-Type"), "application/json") {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "invalid Content-Type (requires application/json)",
		})
	}

	if err := json.Unmarshal(ctx.Body(), &requestBody); err != nil {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "unable to parse request body",
		})
	}

	if requestBody.Date.After(time.Now()) {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "Invalid date: occurs in the future",
		})
	}

	requestBody.Date = requestBody.Date.UTC()

	if !locationsRegexp.MatchString(requestBody.Route) {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "Invalid route format",
		})
	}

	var (
		needsServiceUID = time.Now().UTC().Truncate(24*time.Hour) != requestBody.Date.Truncate(24*time.Hour)
		locations       []string
		services        []string
	)

	{
		lines := strings.Split(requestBody.Route, "\n")
		for i, line := range lines {
			p := strings.Split(line, ",")
			if len(p) == 1 {
				if needsServiceUID && i != len(lines)-1 && requestBody.ManualDistance == 0 {
					ctx.Status(400)
					return ctx.JSON(StockResponse{
						Ok:      false,
						Message: "Service UIDs required as services were run on a different day to today",
					})
				}
				services = append(services, "")
			} else {
				services = append(services, strings.TrimSpace(p[1]))
			}
			locations = append(locations, strings.TrimSpace(p[0]))
		}
	}

	var dist float32
	if requestBody.ManualDistance != 0 {
		dist = requestBody.ManualDistance
	} else {
		var err error
		dist, err = hs.core.GetRouteDistance(locations, services, requestBody.Date)
		if err != nil {
			ctx.Status(400)
			return ctx.JSON(StockResponse{
				Ok:      false,
				Message: "Unable to fetch distance: " + err.Error(),
			})
		}
	}

	var via []string
	if len(locations) > 2 {
		via = locations[1 : len(locations)-1]
	}

	j := &db.Journey{
		ID:   uuid.New(),
		From: &db.StationName{Shortcode: locations[0]},
		To:   &db.StationName{Shortcode: locations[len(locations)-1]},
		Via: util.Map(via, func(x string) *db.StationName {
			return &db.StationName{Shortcode: x}
		}),
		Distance: dist,
		Date:     requestBody.Date,
		Return:   requestBody.IsReturn,
	}

	if err := hs.core.InsertJourney(j); err != nil {
		return fmt.Errorf("inserting new journey: %w", err)
	}

	response.ID = j.ID

	return ctx.JSON(&response)
}
