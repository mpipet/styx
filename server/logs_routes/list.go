package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
)

func (lr *LogsRouter) ListHandler(w http.ResponseWriter, r *http.Request) {

	entries := make(api.ListLogsResponse, 0)

	managedLogs := lr.manager.ListLogs()

	for _, ml := range managedLogs {

		logInfo := ml.Stat()
		entries = append(entries, api.LogInfo(logInfo))
	}

	api.WriteResponse(w, http.StatusOK, entries)
}
