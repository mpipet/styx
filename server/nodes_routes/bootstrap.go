package nodes_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
)

func (nr *NodesRouter) BootstrapHandler(w http.ResponseWriter, r *http.Request) {

	err := nr.nodeManager.Bootstrap()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	api.WriteResponse(w, http.StatusOK, nil)
}

