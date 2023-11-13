package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DistanceWithRoute struct {
	Distance float32
	Route    []string
}

func (dwr *DistanceWithRoute) Add(dw2 *DistanceWithRoute) {
	dwr.Distance += dw2.Distance
	dwr.Route = append(dwr.Route, dw2.Route...)
}

type RTTService struct {
	ServiceUID     string `json:"serviceUid"`
	Headcode       string `json:"runningIdentity"`
	ATOCName       string `json:"atocName"`
	ATOCCode       string `json:"atocCode"`
	IsPassenger    bool   `json:"isPassenger"`
	RunDate        string `json:"runDate"`
	LocationDetail struct {
		DisplayAs           string `json:"displayAs"`
		GBTTBookedDeparture string `json:"gbttBookedDeparture"`
		Destination         []*struct {
			Tiploc      string `json:"tiploc"`
			Description string `json:"description"`
		} `json:"destination"`
	} `json:"locationDetail"`
}

func (c *Core) SearchForServices(from, to, aroundTime string) ([]*RTTService, error) {
	nowUTC := time.Now().UTC()
	today := nowUTC.Format("2006-01-02")

	year := strconv.Itoa(nowUTC.Year())
	month := strconv.Itoa(int(nowUTC.Month()))
	if len(month) == 1 {
		month = "0" + month
	}
	day := strconv.Itoa(nowUTC.Day())
	if len(day) == 1 {
		day = "0" + day
	}

	var rttResp struct {
		Services []*RTTService `json:"services"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	path := fmt.Sprintf("/api/v1/json/search/%s/to/%s/%s/%s/%s", from, to, year, month, day)
	if _, e := strconv.Atoi(aroundTime); aroundTime != "" && e == nil {
		path += "/" + aroundTime
	}
	err := requests.
		URL("https://api.rtt.io").
		Path(path).
		ToJSON(&rttResp).
		BasicAuth(c.config.RealTimeTrains.Username, c.config.RealTimeTrains.Password).
		Fetch(ctx)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("search for service %s->%s: %w", from, to, err)
	}

	n := 0
	for i, service := range rttResp.Services {
		if i == 15 {
			break
		}
		if service.RunDate == today && // If this train started on a different date and runs through midnight
			!strings.EqualFold(service.LocationDetail.DisplayAs, "CANCELLED_CALL") && // If this train was cancelled
			service.IsPassenger {
			rttResp.Services[n] = service
			n += 1
		}
	}

	rttResp.Services = rttResp.Services[:n]

	if len(rttResp.Services) == 0 {
		return nil, errors.New("no route found")
	}

	return rttResp.Services, nil
}

func (c *Core) GetRouteDistance(stations []string, inputServices []string, date time.Time, statusChan chan *util.SSEItem) (*DistanceWithRoute, error) {
	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	if len(month) == 1 {
		month = "0" + month
	}
	day := strconv.Itoa(time.Now().Day())
	if len(day) == 1 {
		day = "0" + day
	}
	todayRunDate := date.Format("2006-01-02")

	var services [][]string
	for _, x := range inputServices {
		if x != "" {
			services = append(services, []string{x})
		} else {
			services = append(services, nil)
		}
	}

	for i := 0; i < len(stations)-1; i += 1 {
		if len(services[i]) == 0 {
			var rttResp struct {
				Services []*RTTService `json:"services"`
			}
			util.SendSSE(statusChan, "status", fmt.Sprintf("Searching for services for leg %s->%s", stations[i], stations[i+1]))
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			err := requests.
				URL("https://api.rtt.io").
				Pathf("/api/v1/json/search/%s/to/%s/%s/%s/%s", stations[i], stations[i+1], year, month, day).
				ToJSON(&rttResp).
				BasicAuth(c.config.RealTimeTrains.Username, c.config.RealTimeTrains.Password).
				Fetch(ctx)
			cancel()
			if err != nil {
				return nil, fmt.Errorf("search for service %s->%s: %w", stations[i], stations[i+1], err)
			}

			var possibleUIDs []string
			for _, service := range rttResp.Services {
				if len(possibleUIDs) == 10 {
					break
				}
				if service.RunDate == todayRunDate && // If this train started on a different date and runs through midnight
					!strings.EqualFold(service.LocationDetail.DisplayAs, "CANCELLED_CALL") && // If this train was cancelled
					service.IsPassenger {
					possibleUIDs = append(possibleUIDs, service.ServiceUID)
				}
			}
			if len(possibleUIDs) == 0 {
				return nil, errors.New("no route found")
			}
			services[i] = possibleUIDs
		}
	}

	var total DistanceWithRoute
	for i := 0; i < len(stations)-1; i += 1 {
		if i != 0 {
			total.Route = append(total.Route, stations[i])
		}
		var dist *DistanceWithRoute
		for _, serv := range services[i] {
			util.SendSSE(statusChan, "status", fmt.Sprintf("Fetching distance for service %s (for leg %s->%s)", serv, stations[i], stations[i+1]))

			d, err := c.getSingleTrainDistance(serv, stations[i], stations[i+1], date)
			if err != nil {
				if !errors.Is(err, noDistancesError) {
					return nil, util.Wrap(err, "scraping train")
				}
				continue
			}
			dist = d
			break
		}

		if dist == nil {
			return nil, util.UserError(fmt.Errorf("no distance information provided for %s -> %s (tried %s) - manual distance required", stations[i], stations[i+1], strings.Join(services[i], ", ")))
		}

		total.Add(dist)
	}

	return &total, nil
}

var shortcodeRegexp = regexp.MustCompile(`[A-Z]{3}`)

var noDistancesError = errors.New("no distances available")

func (c *Core) getSingleTrainDistance(uid, departure, destination string, date time.Time) (*DistanceWithRoute, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var htmlContent string
	err := requests.
		URL("https://www.realtimetrains.co.uk").
		Pathf("/service/gb-nr:%s/%s/detailed", uid, date.Format("2006-01-02")).
		ToString(&htmlContent).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch train with UID %s: %w", uid, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("load RTT HTML: %w", err)
	}

	var waypoints [][3]string

	doc.Find(".location.call,.location.pass").Each(func(i int, selection *goquery.Selection) {
		shortcode := shortcodeRegexp.FindString(
			selection.Find(".location a").Text(),
		)

		if shortcode == "" {
			return
		}

		waypoints = append(waypoints, [3]string{
			shortcode,
			strings.TrimSpace(selection.Find("span.miles").Text()),
			strings.TrimSpace(selection.Find("span.chains").Text()),
		})
	})

	var distances []float32

	for _, wp := range waypoints {
		if strings.EqualFold(wp[0], departure) || strings.EqualFold(wp[0], destination) {
			if wp[1] == "" || wp[2] == "" {
				return nil, noDistancesError
				//return nil, util.UserError(fmt.Errorf("no distance information provided for %s -> %s (%s) - manual distance required", departure, destination, uid))
			}
			miles, err := strconv.Atoi(wp[1])
			if err != nil {
				return nil, fmt.Errorf("parse miles: %w (%#v)", err, wp[1])
			}
			chains, err := strconv.Atoi(wp[2])
			if err != nil {
				return nil, fmt.Errorf("parse chains: %w (%#v)", err, wp[2])
			}
			distances = append(distances, float32(miles)+util.ChainsToMiles(chains))
		}
	}

	if len(distances) != 2 {
		return nil, fmt.Errorf("unexpected number of occurences of source/dest stations in RTT HTML (got %d, expected 2)", len(distances))
	}

	var route []string
	{
		var inbetweenTerminii bool
		for _, wp := range waypoints {
			if strings.EqualFold(wp[0], departure) {
				inbetweenTerminii = true
			} else if strings.EqualFold(wp[0], destination) {
				if !inbetweenTerminii {
					return nil, errors.New("unexpectedly formatted route: destination before departure")
				}
				break
			} else if inbetweenTerminii {
				route = append(route, wp[0])
			}
		}
	}

	return &DistanceWithRoute{
		Distance: float32(math.Abs(float64(distances[1] - distances[0]))),
		Route:    route,
	}, nil
}
