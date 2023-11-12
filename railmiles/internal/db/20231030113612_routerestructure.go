package db

import (
	"context"
	"errors"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewRaw(`CREATE TABLE "railmiles_routes_neo" (
					"from" VARCHAR COLLATE NOCASE, 
					"to" VARCHAR COLLATE NOCASE,
					"sequence" INTEGER, 
					"station" VARCHAR, 
					PRIMARY KEY ("from", "to", "sequence")
				)
			`).Exec(ctx)
			if err != nil {
				return util.Wrap(err, "creating new routes table")
			}

			var existingRoutes []*routeV1
			if err := db.NewSelect().Model(&existingRoutes).Scan(ctx, &existingRoutes); err != nil {
				return util.Wrap(err, "read old routes")
			}

			if len(existingRoutes) != 0 {
				var newRoutes []*routeV2
				for _, oldRoute := range existingRoutes {
					r := &routeV2{
						From: oldRoute.From,
						To:   oldRoute.To,
					}
					for i, point := range oldRoute.Route {
						rq := *r
						rq.Sequence = i
						rq.Station = point
						newRoutes = append(newRoutes, &rq)
					}
				}

				if _, err := db.NewInsert().Model(&newRoutes).ModelTableExpr("railmiles_routes_neo").Exec(ctx); err != nil {
					return util.Wrap(err, "insert new routes")
				}
			}

			if _, err := db.NewRaw(`DROP TABLE "railmiles_routes";`).Exec(ctx); err != nil {
				return util.Wrap(err, "delete old table")
			}

			if _, err := db.NewRaw(`ALTER TABLE "railmiles_routes_neo" RENAME TO "railmiles_routes";`).Exec(ctx); err != nil {
				return util.Wrap(err, "rename new table")
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			return errors.New("not supported")
		},
	)
}
