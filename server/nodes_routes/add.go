package nodes_routes

import (
	"fmt"
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/nodeman"

	"github.com/hashicorp/raft"
)

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
