package core

import (
	"encoding/json"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/util"
)

func (c *Core) GenerateJourneyGeoJSON(journeys []*db.Journey, includeIntermediaries bool) string {
	var stations [][2]string
	{
		for _, journey := range journeys {
			stations = append(stations, [2]string{journey.To.Shortcode, ""}, [2]string{journey.From.Shortcode, ""})
			if includeIntermediaries {
				for _, sn := range journey.Via {
					stations = append(stations, [2]string{sn.Shortcode, "intermediary"})
				}
			}
		}
		stations = util.Deduplicate(stations)
	}

	var res []any

journeyLoop:
	for _, journey := range journeys {
		feature := make(map[string]any)
		feature["type"] = "LineString"
		feature["properties"] = map[string]string{"id": journey.ID.String()}

		routeStations := []string{journey.From.Shortcode}
		routeStations = append(routeStations, util.Map(journey.Via, func(x *db.StationName) string {
			return x.Shortcode
		})...)
		routeStations = append(routeStations, journey.To.Shortcode)

		route, _ := c.GetCallingPoints(journey.ID)
		route = append([]string{journey.From.Shortcode}, route...)
		route = append(route, journey.To.Shortcode)

		route = append(route, routeStations[len(routeStations)-1])

		var coords [][]float32
		{
			last := len(route) - 1
			for i, point := range route {
				details := GetStationDetail(point)
				if details == nil {
					if i == 0 || i == last {
						continue journeyLoop
					}
					continue
				}
				coords = append(coords, []float32{details.Lon, details.Lat})
			}
		}

		feature["coordinates"] = coords
		res = append(res, feature)
	}

	for _, station := range stations {
		stationDetails := GetStationDetail(station[0])
		if stationDetails == nil {
			continue
		}

		props := map[string]any{"name": station[0] + " " + GetStationName(station[0])}

		if station[1] != "" {
			props["type"] = station[1]
		}

		res = append(res, map[string]any{
			"type":       "Feature",
			"properties": props,
			"geometry":   map[string]any{"type": "Point", "coordinates": []float32{stationDetails.Lon, stationDetails.Lat}},
		})
	}

	o, _ := json.Marshal(res)
	return string(o)
}
