package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/server/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type LogsRouter struct {
	router        *mux.Router
	manager       *logman.LogManager
	Config        config.Config
	schemaDecoder *schema.Decoder
}

func RegisterRoutes(router *mux.Router, logManager *logman.LogManager, config config.Config) (lr *LogsRouter) {

	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	lr = &LogsRouter{
		router:        router,
		manager:       logManager,
		Config:        config,
		schemaDecoder: decoder,
	}

	router.HandleFunc("", lr.ListHandler).
		Methods(http.MethodGet)

	router.HandleFunc("", lr.CreateHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}", lr.GetHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/{name}", lr.DeleteHandler).
		Methods(http.MethodDelete)

	router.HandleFunc("/{name}/backup", lr.BackupHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/restore", lr.RestoreHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/records/batch", lr.WriteBatchHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/records/batch", lr.ReadBatchHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/{name}/records", lr.WriteTCPHandler).
		Methods(http.MethodPost).
		Headers("Upgrade", "tcp")

	router.HandleFunc("/{name}/records", lr.ReadTCPHandler).
		Methods(http.MethodGet).
		Headers("Upgrade", "tcp")

	router.HandleFunc("/{name}/records", lr.WriteHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/records", lr.ReadHandler).
		Methods(http.MethodGet)

	return lr
}
