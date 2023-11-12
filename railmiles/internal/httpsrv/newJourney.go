package httpsrv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"strings"
	"time"
)

type newJourneyRequest struct {
	Date           time.Time  `json:"date"`
	Route          [][]string `json:"route"`
	ManualDistance float32    `json:"manualDistance"`
	IsReturn       bool       `json:"isReturn"`
}

func (hs *httpServer) newJourney(ctx *fiber.Ctx) error {
	if !strings.EqualFold(ctx.Get("Content-Type"), "application/json") {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "invalid Content-Type (requires application/json)",
		})
	}

	requestBody := new(newJourneyRequest)

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

	var (
		needsServiceUID = time.Now().UTC().Truncate(24*time.Hour) != requestBody.Date.Truncate(24*time.Hour)
		locations       []string
		services        []string
	)

	{
		for i, line := range requestBody.Route {
			if line[1] == "" {
				if needsServiceUID && i != len(requestBody.Route)-1 && requestBody.ManualDistance == 0 {
					ctx.Status(400)
					return ctx.JSON(StockResponse{
						Ok:      false,
						Message: "Service UIDs required as services were run on a different day to today",
					})
				}
				services = append(services, "")
			} else {
				services = append(services, strings.TrimSpace(line[1]))
			}
			locations = append(locations, strings.TrimSpace(line[0]))
		}
	}

	pid, ch := hs.newProcessor()

	go hs.processNewJourney(requestBody, locations, services, pid, ch)

	ctx.Status(202)
	return ctx.JSON(&struct {
		ProcessorID uuid.UUID `json:"processorID"`
	}{pid})
}

func (hs *httpServer) processNewJourney(requestBody *newJourneyRequest, locations, services []string, processID uuid.UUID, output chan *util.SSEItem) {
	var dist *core.DistanceWithRoute
	if requestBody.ManualDistance != 0 {
		dist.Distance = requestBody.ManualDistance
	} else {
		var err error
		dist, err = hs.core.GetRouteDistance(locations, services, requestBody.Date, output)
		if err != nil {
			output <- &util.SSEItem{
				Event:   "error",
				Message: "Unable to fetch distance: " + err.Error(),
			}
			hs.cleanupProcessor(processID)
			return
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
		Distance: dist.Distance,
		Date:     requestBody.Date,
		Return:   requestBody.IsReturn,
	}

	if err := hs.core.InsertJourney(j); err != nil {
		slog.Error("error when inserting new journey", "err", err)
		output <- &util.SSEItem{
			Event:   "error",
			Message: "Internal Server Error",
		}
		hs.cleanupProcessor(processID)
		return
	}

	if err := hs.core.InsertRoute(j.ID, dist.Route); err != nil {
		slog.Error("error when inserting new journey route", "err", err)
		output <- &util.SSEItem{
			Event:   "error",
			Message: "Internal Server Error",
		}
		hs.cleanupProcessor(processID)
		return
	}

	output <- &util.SSEItem{
		Event:   "finished",
		Message: j.ID.String(),
	}
	hs.cleanupProcessor(processID)
}

func (hs *httpServer) serveProcessorStream(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Println(err)
		return fiber.ErrNotFound
	}

	hs.journeyProcessorLock.Lock()
	channel, found := hs.journeyProcessors[id]
	hs.journeyProcessorLock.Unlock()

	if !found {
		fmt.Println("no lol")
		return fiber.ErrNotFound
	}

	ctx.Set("Content-Type", "text/event-stream")
	fr := ctx.Response()
	fr.SetBodyStreamWriter(func(w *bufio.Writer) {
		for item := range channel {
			_, _ = w.Write([]byte(item.String()))
			if err := w.Flush(); err != nil {
				// client disconnected
				return
			}
		}
	})
	return nil
}
