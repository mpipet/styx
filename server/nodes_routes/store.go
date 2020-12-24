package nodes_routes

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/nodeman"

	"github.com/gorilla/mux"
)

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
