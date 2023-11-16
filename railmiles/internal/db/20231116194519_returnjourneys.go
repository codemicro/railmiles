package db

import (
	"context"
	"errors"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slices"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			var returns []*journeyV1

			if err := db.NewSelect().
				Model(&returns).
				Where("return = true").
				Scan(context.Background()); err != nil {
				return util.Wrap(err, "querying return journeys")
			}

			if _, err := db.NewRaw(`ALTER TABLE "railmiles_journeys" ADD COLUMN "return_id" uuid;`).Exec(ctx); err != nil {
				return util.Wrap(err, "adding return id column to routes table")
			}

			if _, err := db.NewRaw(`ALTER TABLE "railmiles_journeys" DROP COLUMN "return";`).Exec(ctx); err != nil {
				return util.Wrap(err, "dropping return column")
			}

			if _, err := db.NewRaw(`ALTER TABLE "railmiles_journeys" RENAME TO "railmiles_journeys_v2";`).Exec(ctx); err != nil {
				return util.Wrap(err, "dropping return column")
			}

			for _, outbound := range returns {
				var n []*StationName
				if len(outbound.Via) != 0 {
					// This ensures that empty via entries remain as `nil` ensuring that it gets inserted into the DB as
					// null correctly.
					n = make([]*StationName, len(outbound.Via))
					copy(n, outbound.Via)
					slices.Reverse(n)
				}

				j := Journey{
					ID:       uuid.New(),
					From:     outbound.To,
					To:       outbound.From,
					Via:      n,
					Distance: outbound.Distance,
					Date:     outbound.Date,
					ReturnID: &outbound.ID,
				}

				if _, err := db.NewInsert().Model(&j).Exec(ctx); err != nil {
					return util.Wrap(err, "insert new journey")
				}

				var route []string
				err := db.NewSelect().Model((*Route)(nil)).Column("station").Where(`journey_id = ?`, outbound.ID).Order("sequence").Scan(context.Background(), &route)
				if err != nil {
					return util.Wrap(err, "get route")
				}

				if len(route) != 0 {
					slices.Reverse(route)

					var routeParts []*Route
					r := &Route{
						JourneyID: j.ID,
					}
					for i, point := range route {
						rq := *r
						rq.Sequence = i
						rq.Station = point
						routeParts = append(routeParts, &rq)
					}
					_, err = db.NewInsert().Model(&routeParts).Exec(context.Background())
					if err != nil {
						return util.Wrap(err, "insert new route")
					}
				}

				db.NewUpdate().Model((*Journey)(nil)).Set("return_id = ?", j.ID).Where("id = ?", outbound.ID).Exec(context.Background())
				if err != nil {
					return util.Wrap(err, "update old route with return ID")
				}
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			return errors.New("not supported")
		},
	)
}
