package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/rs/zerolog/log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (c *Core) GetRouteDistance(stations []string, services []string, date time.Time) (float32, error) {
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

	for i := 0; i < len(stations)-1; i += 1 {
		if services[i] == "" {
			var rttResp struct {
				Services []struct {
					ServiceUid     string `json:"serviceUid"`
					IsPassenger    bool   `json:"isPassenger"`
					RunDate        string `json:"runDate"`
					LocationDetail struct {
						DisplayAs string `json:"displayAs"`
					} `json:"locationDetail"`
				} `json:"services"`
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			err := requests.
				URL("https://api.rtt.io").
				Pathf("/api/v1/json/search/%s/to/%s/%s/%s/%s", stations[i], stations[i+1], year, month, day).
				ToJSON(&rttResp).
				BasicAuth(c.config.RealTimeTrains.Username, c.config.RealTimeTrains.Password).
				Fetch(ctx)
			cancel()
			if err != nil {
				return 0, fmt.Errorf("search for service %s->%s: %w", stations[i], stations[i+1], err)
			}

			uid := ""
			for _, service := range rttResp.Services {
				if service.RunDate == todayRunDate && // If this train started on a different date and runs through midnight
					!strings.EqualFold(service.LocationDetail.DisplayAs, "CANCELLED_CALL") && // If this train was cancelled
					service.IsPassenger {
					uid = service.ServiceUid
					break
				}
			}
			if uid == "" {
				return 0, errors.New("no route found")
			}
			services[i] = uid
		}
	}

	var total float32
	for i := 0; i < len(stations)-1; i += 1 {
		dist, err := c.getSingleTrainDistance(services[i], stations[i], stations[i+1], date)
		if err != nil {
			return 0, util.Wrap(err, "scraping train")
		}
		total += dist
	}

	return total, nil
}

var shortcodeRegexp = regexp.MustCompile(`[A-Z]{3}`)

func (c *Core) getSingleTrainDistance(uid, departure, destination string, date time.Time) (float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var htmlContent string
	err := requests.
		URL("https://www.realtimetrains.co.uk").
		Pathf("/service/gb-nr:%s/%s/detailed", uid, date.Format("2006-01-02")).
		ToString(&htmlContent).
		Fetch(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch train with UID %s: %w", uid, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(htmlContent))
	if err != nil {
		return 0, fmt.Errorf("load RTT HTML: %w", err)
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
				return 0, util.UserError(fmt.Errorf("no distance information provided for %s -> %s (%s) - manual distance required", departure, destination, uid))
			}
			miles, err := strconv.Atoi(wp[1])
			if err != nil {
				return 0, fmt.Errorf("parse miles: %w (%#v)", err, wp[1])
			}
			chains, err := strconv.Atoi(wp[2])
			if err != nil {
				return 0, fmt.Errorf("parse chains: %w (%#v)", err, wp[2])
			}
			distances = append(distances, float32(miles)+util.ChainsToMiles(chains))
		}
	}

	if len(distances) != 2 {
		return 0, fmt.Errorf("unexpected number of occurences of source/dest stations in RTT HTML (got %d, expected 2)", len(distances))
	}

	var route []string
	{
		var inbetweenTerminii bool
		for _, wp := range waypoints {
			if strings.EqualFold(wp[0], departure) {
				inbetweenTerminii = true
			} else if strings.EqualFold(wp[0], destination) {
				if !inbetweenTerminii {
					return 0, errors.New("unexpectedly formatted route: destination before departure")
				}
				break
			} else if inbetweenTerminii {
				route = append(route, wp[0])
			}
		}
	}

	err = c.InsertRoute(&db.Route{
		From:  departure,
		To:    destination,
		Route: route,
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to save route")
	}

	return float32(math.Abs(float64(distances[1] - distances[0]))), nil
}
