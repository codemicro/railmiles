package db

import (
	"context"
	"database/sql"
	"github.com/codemicro/railmiles/railmiles/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
	"golang.org/x/exp/slog"
	"time"
)

var Migrations = migrate.NewMigrations()

type DB struct {
	DB *bun.DB
}

func New(conf *config.Config) (*DB, error) {
	dsn := conf.Database.DSN
	slog.Info("connecting to database")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // https://github.com/mattn/go-sqlite3/issues/274#issuecomment-191597862

	b := bun.NewDB(db, sqlitedialect.New())

	if conf.Debug {
		b.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return &DB{b}, nil
}

func (db *DB) Migrate() error {
	slog.Info("running database migrations")

	mig := migrate.NewMigrator(db.DB, Migrations)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := mig.Init(ctx); err != nil {
		return err
	}

	group, err := mig.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		slog.Info("no migrations applied (database up to date)")
	} else {
		slog.Info("migrations applied")
	}

	return nil
}
