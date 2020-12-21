package nodes_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/nodeman"
	"gitlab.com/dataptive/styx/server/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/hashicorp/raft"
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

	router.HandleFunc("", nr.ListHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/bootstrap", nr.BootstrapHandler).
		Methods(http.MethodPost)

	return nr
}

func (nr *NodesRouter) BootstrapHandler(w http.ResponseWriter, r *http.Request) {

	err := nr.nodeManager.Bootstrap()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	api.WriteResponse(w, http.StatusOK, nil)
}

func (nr *NodesRouter) ListHandler(w http.ResponseWriter, r *http.Request) {

	entries := api.ListNodesResponse{}

	raftNodes, err := nr.nodeManager.ListNodes()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	for _, rn := range raftNodes {

		var state string
		var suffrage string

		if rn.State == raft.Leader {
			state = "leader"
		} else {
			state = "follower"
		}

		switch rn.Suffrage {
		case raft.Voter:
			suffrage = "voter"
		case raft.Nonvoter:
			suffrage = "nonvoter"
		case raft.Staging:
			suffrage = "staging"
		}

		node := api.Node{
			Name:     string(rn.Name),
			State:    state,
			Suffrage: suffrage,
			Address:  string(rn.Address),
		}

		entries = append(entries, node)
	}

	api.WriteResponse(w, http.StatusOK, entries)
}
