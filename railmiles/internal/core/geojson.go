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

		if len(route) == 0 && len(journey.Via) != 0 {
			// This likely means that the journey had calling points listed as well as a manual distance, hence no auto
			// route was inserted into the database. This check allows us to fit the line to the calling points so we
			// don't end up with a line that goes direct between A and B without passing through C or D.
			route = make([]string, len(journey.Via))
			for i, x := range journey.Via {
				route[i] = x.Shortcode
			}
		}

		route = append([]string{journey.From.Shortcode}, route...)
		route = append(route, journey.To.Shortcode)

		route = append(route, routeStations[len(routeStations)-1])

		var coords [][2]float32
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
				coords = append(coords, [2]float32{details.Lon, details.Lat})
			}
		}

		feature["coordinates"] = smoothLine(coords)
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

func smoothLine(coords [][2]float32) [][2]float32 {
	// Chaikin’s curve algorithm

	const scale = 0.125

	for range 5 {
		newPoints := [][2]float32{coords[0]}
		for i, point := range coords[1 : len(coords)-1] {
			i += 1 // since we're skipping the first

			previousPoint := coords[i-1]
			nextPoint := coords[i+1]

			{
				dx := point[0] - previousPoint[0]
				dy := point[1] - previousPoint[1]

				newPoints = append(newPoints, [2]float32{point[0] - (dx * scale), point[1] - (dy * scale)})
			}

			{
				dx := nextPoint[0] - point[0]
				dy := nextPoint[1] - point[1]

				newPoints = append(newPoints, [2]float32{point[0] + (dx * scale), point[1] + (dy * scale)})
			}
		}
		newPoints = append(newPoints, coords[len(coords)-1])
		coords = newPoints
	}
	return coords
}
