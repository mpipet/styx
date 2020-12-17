package logs_routes

import (
	"gitlab.com/dataptive/styx/manager"
	"gitlab.com/dataptive/styx/server/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type LogsRouter struct {
	router        *mux.Router
	manager       *manager.LogManager
	Config        config.Config
	schemaDecoder *schema.Decoder
}

func RegisterRoutes(router *mux.Router, logManager *manager.LogManager, config config.Config) (lr *LogsRouter) {

	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	lr = &LogsRouter{
		router:        router,
		manager:       logManager,
		Config:        config,
		schemaDecoder: decoder,
	}

	return lr
}
