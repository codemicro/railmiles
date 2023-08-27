package httpsrv

import (
	"github.com/codemicro/railmiles/railmiles/internal/config"
	"github.com/codemicro/railmiles/railmiles/internal/core"
	"github.com/codemicro/railmiles/railmiles/internal/util"
	webAssets "github.com/codemicro/railmiles/web"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/google/uuid"
	"net/http"
	"sync"
)

type httpServer struct {
	config *config.Config
	core   *core.Core

	journeyProcessorLock sync.Mutex
	journeyProcessors    map[uuid.UUID]chan *util.SSEItem
}

func Run(conf *config.Config, c *core.Core) error {
	srv := &httpServer{
		config: conf,
		core:   c,

		journeyProcessors: make(map[uuid.UUID]chan *util.SSEItem),
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: !conf.Debug,
	})
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
	app.Get("/api/journeys/processor/:id", hs.serveProcessorStream)
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

func (hs *httpServer) newProcessor() (uuid.UUID, chan *util.SSEItem) {
	hs.journeyProcessorLock.Lock()
	id := uuid.New()
	ch := make(chan *util.SSEItem, 16)
	hs.journeyProcessors[id] = ch
	hs.journeyProcessorLock.Unlock()
	return id, ch
}

func (hs *httpServer) cleanupProcessor(id uuid.UUID) {
	hs.journeyProcessorLock.Lock()
	ch, found := hs.journeyProcessors[id]
	if !found {
		return
	}
	close(ch)
	delete(hs.journeyProcessors, id)
	hs.journeyProcessorLock.Unlock()
}
