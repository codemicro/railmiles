package db

import (
	"context"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewRaw(`CREATE TABLE "railmiles_routes" (
					"from" VARCHAR COLLATE NOCASE, 
					"to" VARCHAR COLLATE NOCASE, 
					"route" VARCHAR, 
					PRIMARY KEY ("from", "to")
				)
			`).Exec(ctx)
			if err != nil {
				return util.Wrap(err, "creating routes table")
			}
			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewDropTable().Model((*Route)(nil)).Exec(ctx); err != nil {
				return util.Wrap(err, "dropping routes table")
			}
			return nil
		},
	)
}
