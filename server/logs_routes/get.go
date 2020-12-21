package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) GetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	managedLog, err := lr.manager.GetLog(name)
	if err == logman.ErrNotExist {
		api.WriteError(w, http.StatusNotFound, api.ErrLogNotFound)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	logInfo := managedLog.Stat()

	api.WriteResponse(w, http.StatusOK, api.GetLogResponse(logInfo))
}
