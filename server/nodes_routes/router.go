package nodes_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/nodeman"
	"gitlab.com/dataptive/styx/server/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type NodesRouter struct {
	router        *mux.Router
	nodeManager   *nodeman.NodeManager
	config        config.Config
	schemaDecoder *schema.Decoder
}

func RegisterRoutes(router *mux.Router, nodeManager *nodeman.NodeManager, config config.Config) (nr *NodesRouter) {

	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	nr = &NodesRouter{
		router:        router,
		nodeManager:   nodeManager,
		config:        config,
		schemaDecoder: decoder,
	}

	router.HandleFunc("", nr.RaftHandler).
		Methods(http.MethodGet).
		Headers("Upgrade", api.RaftProtocolString)

	router.HandleFunc("", nr.ListHandler).
		Methods(http.MethodGet)

	router.HandleFunc("", nr.AddNodeHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}", nr.RemoveNodeHandler).
		Methods(http.MethodDelete)

	router.HandleFunc("/bootstrap", nr.BootstrapHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/store/{key}", nr.StoreGetHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/store/{key}", nr.StoreSetHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/store/{key}", nr.StoreDeleteHandler).
		Methods(http.MethodDelete)

	router.HandleFunc("/store", nr.StoreListHandler).
		Methods(http.MethodGet)

	return nr
}
