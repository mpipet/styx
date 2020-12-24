package nodes_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"

	"github.com/hashicorp/raft"
)

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
