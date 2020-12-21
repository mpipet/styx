package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) DeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	err := lr.manager.DeleteLog(name)
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

	api.WriteResponse(w, http.StatusOK, nil)
}
