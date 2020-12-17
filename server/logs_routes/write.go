package logs_routes

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/manager"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) WriteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	contentLength := r.Header.Get("Content-Length")

	if contentLength == "" {
		api.WriteError(w, http.StatusBadRequest, api.ErrMissingContentLength)
		logger.Debug(nil)
		return
	}

	recordSize, err := strconv.Atoi(contentLength)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	if recordSize == 0 {
		api.WriteResponse(w, http.StatusOK, api.WriteRecordResponse{})
		return
	}

	managedLog, err := lr.manager.GetLog(name)
	if err == manager.ErrNotExist {
		api.WriteError(w, http.StatusNotFound, api.ErrLogNotFound)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	logWriter, err := managedLog.NewWriter(recio.ModeAuto)
	if err == manager.ErrUnavailable {
		api.WriteError(w, http.StatusBadRequest, api.ErrLogNotAvailable)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	var progress log.SyncProgress

	logWriter.HandleSync(func(syncProgress log.SyncProgress) {
		progress = syncProgress
	})

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	record := log.Record(payload)

	_, err = logWriter.Write(&record)
	if err != nil {
		logWriter.Close()
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = logWriter.Flush()
	if err != nil {
		logWriter.Close()
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = logWriter.Close()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	response := api.WriteRecordResponse(progress)

	api.WriteResponse(w, http.StatusOK, response)
}
