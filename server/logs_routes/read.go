package logs_routes

import (
	"io"
	"net/http"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/manager"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) ReadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	params := api.ReadRecordParams{
		Whence:   log.SeekOrigin,
		Position: 0,
	}
	query := r.URL.Query()

	err := lr.schemaDecoder.Decode(&params, query)
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
		return
	}

	err = params.Validate()
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
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

	logReader, err := managedLog.NewReader(lr.Config.HTTPReadBufferSize, false, recio.ModeAuto)
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

	err = logReader.Seek(params.Position, params.Whence)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		logReader.Close()
		return
	}

	record := log.Record{}

	_, err = logReader.Read(&record)
	if err == io.EOF {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		logReader.Close()
		return
	}

	err = logReader.Close()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(record)))
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(record)
	if err != nil {
		logger.Debug(err)
		return
	}
}
