package db

import (
	"context"
	"errors"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewRaw(`CREATE TABLE "railmiles_routes_v3" (
					"journey_id" uuid,
					"sequence" INTEGER,
					"station" VARCHAR,
					PRIMARY KEY ("journey_id", "sequence")
				)
			`).Exec(ctx)
			if err != nil {
				return util.Wrap(err, "creating routes table")
			}

			var existingRoutes []*routeV2
			if err := db.NewSelect().Model(&existingRoutes).Scan(ctx, &existingRoutes); err != nil {
				return util.Wrap(err, "read old routes")
			}

			routeMap := make(map[[2]string][]*routeV2)
			for _, x := range existingRoutes {
				routeMap[[2]string{strings.ToUpper(x.From), strings.ToUpper(x.To)}] = append(routeMap[[2]string{strings.ToUpper(x.From), strings.ToUpper(x.To)}], x)
			}

			for _, l := range routeMap {
				sort.Slice(l, func(i, j int) bool {
					return l[i].Sequence < l[j].Sequence
				})
			}

			var journeys []*Journey
			if err := db.NewSelect().Model(&journeys).Scan(ctx, &journeys); err != nil {
				return util.Wrap(err, "read journeys")
			}

			var toInsert [][]any

			for _, journey := range journeys {
				stops := []string{journey.From.Shortcode}
				stops = append(stops, util.Map(journey.Via, func(t *StationName) string {
					return t.Shortcode
				})...)
				stops = append(stops, journey.To.Shortcode)

				var overallRoute []string

				for i := 0; i < len(stops)-1; i += 1 {
					from := strings.ToUpper(stops[i])
					to := strings.ToUpper(stops[i+1])

					overallRoute = append(overallRoute, from)

					t := routeMap[[2]string{from, to}]
					if len(t) == 0 {
						t = routeMap[[2]string{to, from}]
						slices.Reverse(t)
					}
					for _, x := range t {
						overallRoute = append(overallRoute, x.Station)
					}
				}

				overallRoute = overallRoute[1:]

				for i, x := range overallRoute {
					toInsert = append(toInsert, []any{journey.ID, i, x})
				}
			}

			if len(toInsert) != 0 {
				// https://stackoverflow.com/questions/15858466/limit-on-multiple-rows-insert
				var (
					batches [][][]any
					acc     [][]any
				)
				for i := 0; i < len(toInsert); i += 1 {
					acc = append(acc, toInsert[i])
					if i%50 == 0 {
						batches = append(batches, acc)
						acc = nil
					}
				}
				if len(acc) != 0 {
					batches = append(batches, acc)
				}

				for _, batch := range batches {
					var ins []any
					for _, x := range batch {
						ins = append(ins, x...)
					}

					queryStr := ` `
					for i := 0; i < len(batch); i += 1 {
						queryStr += "(?, ?, ?), "
					}
					queryStr = queryStr[:len(queryStr)-2] + ";"
					if _, err := db.NewRaw(`INSERT INTO railmiles_routes_v3 ("journey_id", "sequence", "station") VALUES `+queryStr, ins...).Exec(ctx); err != nil {
						return util.Wrap(err, "insert journey")
					}
				}

			}

			if _, err := db.NewRaw(`DROP TABLE "railmiles_routes";`).Exec(ctx); err != nil {
				return util.Wrap(err, "delete old table")
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			return errors.New("not supported")
		},
	)
}
