package nodes_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/logger"
)

func (nr *NodesRouter) RaftHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := UpgradeTCP(w)
	if err != nil {
		logger.Debug(err)
		return
	}

	nr.nodeManager.AcceptHandler(conn)
}
