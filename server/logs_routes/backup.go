package logs_routes

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) BackupHandler(w http.ResponseWriter, r *http.Request) {

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

	filename := fmt.Sprintf("%s-%d.tar.gz", name, time.Now().Unix())
	attachment := fmt.Sprintf("attachment; filename=%s", filename)

	w.Header().Set("Content-Disposition", attachment)
	w.Header().Set("Content-Type", "application/gzip")

	w.WriteHeader(200)

	err = managedLog.Backup(w)
	if err != nil {
		logger.Debug(err)
		return
	}
}
