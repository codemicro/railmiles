package db

import (
	"context"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewCreateTable().Model((*Journey)(nil)).Exec(ctx); err != nil {
				return util.Wrap(err, "creating journey table")
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.NewDropTable().Model((*Journey)(nil)).Exec(ctx); err != nil {
				return util.Wrap(err, "dropping journey table")
			}

			return nil
		},
	)
}
