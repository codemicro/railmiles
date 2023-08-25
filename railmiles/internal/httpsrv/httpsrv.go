package httpsrv

import (
	"github.com/codemicro/railmiles/railmiles/internal/config"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	webAssets "github.com/codemicro/railmiles/web"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"net/http"
)

type httpServer struct {
	config *config.Config
	core   *core.Core
}

func Run(conf *config.Config, c *core.Core) error {
	srv := &httpServer{
		config: conf,
		core:   c,
	}

	app := fiber.New()
	if conf.Debug {
		app.Use(cors.New())
	}
	srv.registerRoutes(app)

	return app.Listen(conf.HTTPAddress())
}

func (hs *httpServer) registerRoutes(app *fiber.App) {
	app.Get("/api/dashboard", hs.dashboardInfo)
	app.Get("/api/journeys", hs.journeyListing)
	app.Post("/api/journeys", hs.newJourney)
	app.Get("/api/journeys/:id", hs.getJourney)
	app.Delete("/api/journeys/:id", hs.deleteJourney)
	app.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(webAssets.Public),
		PathPrefix: "public",
	}))
}

type StockResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}
