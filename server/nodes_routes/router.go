package nodes_routes

import (
	"fmt"
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

	router.HandleFunc("", nr.RaftHandler).
		Methods(http.MethodGet).
		Headers("Upgrade", nodeman.HashicorpRaftProtocolString)

	router.HandleFunc("", nr.ListHandler).
		Methods(http.MethodGet)

	router.HandleFunc("", nr.AddNodeHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}", nr.RemoveNodeHandler).
		Methods(http.MethodDelete)

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

func (nr *NodesRouter) AddNodeHandler(w http.ResponseWriter, r *http.Request) {

	if !nr.nodeManager.IsLeader() {
		leader := string(nr.nodeManager.Leader())
		location := fmt.Sprintf("http://%s/nodes", leader)
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
		return
	}

	form := api.AddNodeForm{
		Voter: true,
	}

	err := r.ParseForm()
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = nr.schemaDecoder.Decode(&form, r.PostForm)
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
		return
	}

	nodeName := raft.ServerID(form.Name)
	nodeAddr := raft.ServerAddress(form.Address)
	voter := form.Voter

	err = nr.nodeManager.AddNode(nodeName, nodeAddr, voter)

	if err == nodeman.ErrNotLeader {
		leaderAddr := string(nr.nodeManager.Leader())
		location := fmt.Sprintf("http://%s/nodes", leaderAddr)
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	api.WriteResponse(w, http.StatusOK, nil)
}

func (nr *NodesRouter) RemoveNodeHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	nodeName := raft.ServerID(name)

	err := nr.nodeManager.RemoveNode(nodeName)

	if err == nodeman.ErrNotLeader {
		leaderAddr := string(nr.nodeManager.Leader())
		location := fmt.Sprintf("http://%s/nodes/%s", leaderAddr, name)
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
		return
	}

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

		var suffrage string

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
			Leader:   rn.Leader,
			Suffrage: suffrage,
			Address:  string(rn.Address),
		}

		entries = append(entries, node)
	}

	api.WriteResponse(w, http.StatusOK, entries)
}

func (nr *NodesRouter) RaftHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := UpgradeTCP(w)
	if err != nil {
		logger.Debug(err)
		return
	}

	nr.nodeManager.AcceptHandler(conn)
}
