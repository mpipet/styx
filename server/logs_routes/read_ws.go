package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func (lr *LogsRouter) ReadWSHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	vars := mux.Vars(r)
	name := vars["name"]

	params := api.ReadRecordsWSParams{
		Whence:   log.SeekOrigin,
		Position: 0,
	}
	query := r.URL.Query()

	err = lr.schemaDecoder.Decode(&params, query)
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

	logReader, err := managedLog.NewReader(lr.config.HTTPReadBufferSize, true, recio.ModeManual)
	if err == logman.ErrUnavailable {
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

	conn, err := UpgradeWebsocket(w, r, lr.config.CORSAllowedOrigins)
	if err != nil {
		logger.Debug(err)
		logReader.Close()
		return
	}

	err = readWS(conn, logReader)
	if err != nil {
		logger.Debug(err)

		// Close reader to unlock follow
		// if not already done.
		logReader.Close()

		// Close conn in case its still open.
		conn.Close()

		return
	}

	err = logReader.Close()
	if err != nil {
		logger.Debug(err)

		conn.Close()
	}

	err = conn.Close()
	if err != nil {
		logger.Debug(err)
	}
}

func readWS(w *websocket.Conn, lr *log.LogReader) (err error) {

	record := log.Record{}

	for {
		_, err := lr.Read(&record)
		if err == recio.ErrMustFill {

			err = lr.Fill()
			if err != nil {
				return err
			}

			continue
		}

		if err != nil {
			return err
		}

		err = w.WriteMessage(websocket.BinaryMessage, []byte(record))
		if err != nil {
			return err
		}
	}

	return nil
}
