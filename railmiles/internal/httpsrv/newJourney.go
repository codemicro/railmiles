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

type searchServicesRequest [][]string

func (hs *httpServer) searchServices(ctx *fiber.Ctx) error {
	if !strings.EqualFold(ctx.Get("Content-Type"), "application/json") {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "invalid Content-Type (requires application/json)",
		})
	}

	var requestBody searchServicesRequest

	if err := json.Unmarshal(ctx.Body(), &requestBody); err != nil {
		ctx.Status(400)
		return ctx.JSON(StockResponse{
			Ok:      false,
			Message: "unable to parse request body",
		})
	}

	for _, stns := range requestBody {
		if ls := len(stns); !(2 <= ls && ls <= 3) {
			ctx.Status(400)
			return ctx.JSON(StockResponse{
				Ok:      false,
				Message: "incorrect number of stations",
			})
		}
		stns[0] = strings.ToUpper(stns[0])
		stns[1] = strings.ToUpper(stns[1])
	}

	type searchResult struct {
		From     string     `json:"from"`
		To       string     `json:"to"`
		Services [][]string `json:"service"`
	}

	var res []*searchResult

	for _, stationPair := range requestBody {
		around := ""
		if len(stationPair) > 2 {
			around = stationPair[2]
		}
		servs, err := hs.core.SearchForServices(stationPair[0], stationPair[1], around)
		if err != nil {
			return util.Wrap(err, "searching for services")
		}
		x := &searchResult{
			From:     stationPair[0],
			To:       stationPair[1],
			Services: nil,
		}
		for _, serv := range servs {
			dest := "unknown"
			if len(serv.LocationDetail.Destination) > 0 {
				dest = serv.LocationDetail.Destination[0].Description
			}
			x.Services = append(x.Services, []string{
				serv.ServiceUID,
				fmt.Sprintf("%s to %s [%s] (%s %s)", serv.LocationDetail.GBTTBookedDeparture, dest, serv.Headcode, serv.ATOCCode, serv.ATOCName),
			})
		}
		res = append(res, x)
	}

	return ctx.JSON(res)
}

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
