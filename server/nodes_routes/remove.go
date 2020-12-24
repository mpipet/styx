package nodes_routes

import (
	"fmt"
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/nodeman"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
)

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
