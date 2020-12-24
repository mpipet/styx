package nodes_routes

import (
	"fmt"
	"io/ioutil"
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

func (nr *NodesRouter) StoreGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["key"]

	value, err := nr.nodeManager.GetUnsafe(key)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	api.WriteResponse(w, http.StatusOK, value)
}

func (nr *NodesRouter) StoreSetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["key"]

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	value := string(data)

	err = nr.nodeManager.Set(key, value)

	if err == nodeman.ErrNotLeader {
		leaderAddr := string(nr.nodeManager.Leader())
		location := fmt.Sprintf("http://%s/nodes/store/%s", leaderAddr, key)
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

func (nr *NodesRouter) StoreDeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["key"]

	err := nr.nodeManager.Delete(key)

	if err == nodeman.ErrNotLeader {
		leaderAddr := string(nr.nodeManager.Leader())
		location := fmt.Sprintf("http://%s/nodes/store/%s", leaderAddr, key)
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

func (nr *NodesRouter) StoreListHandler(w http.ResponseWriter, r *http.Request) {

	value, err := nr.nodeManager.ListUnsafe()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	api.WriteResponse(w, http.StatusOK, value)
}
