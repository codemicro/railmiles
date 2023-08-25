package main

import (
	"github.com/codemicro/railmiles/railmiles/internal/config"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	"github.com/codemicro/railmiles/railmiles/internal/db"
	"github.com/codemicro/railmiles/railmiles/internal/httpsrv"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	"golang.org/x/exp/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error("unhandled error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	conf, err := config.Load()
	if err != nil {
		return util.Wrap(err, "loading configuration")
	}

	database, err := db.New(conf)
	if err != nil {
		return util.Wrap(err, "opening database")
	}

	if err := database.Migrate(); err != nil {
		return util.Wrap(err, "migrating database")
	}

	c := core.New(conf, database)

	return httpsrv.Run(conf, c)
}
