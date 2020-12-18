package logs_routes

import (
	"io"
	"net/http"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/api/tcp"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/manager"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) WriteTCPHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	vars := mux.Vars(r)
	name := vars["name"]

	remoteTimeout := lr.Config.TCPTimeout

	// TODO: Change the header name to a more a adequate one.
	rawTimeout := r.Header.Get("Request-Timeout")
	if rawTimeout != "" {

		remoteTimeout, err = strconv.Atoi(rawTimeout)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
			logger.Debug(err)
			return
		}
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

	conn, err := UpgradeTCP(w)
	if err != nil {
		logger.Debug(err)
		logWriter.Close()
		return
	}

	err = conn.SetReadBuffer(lr.Config.TCPReadBufferSize)
	if err != nil {
		logger.Warn(err)
	}

	err = conn.SetWriteBuffer(lr.Config.TCPWriteBufferSize)
	if err != nil {
		logger.Warn(err)
	}

	tr := tcp.NewTCPReader(conn, lr.Config.TCPWriteBufferSize, lr.Config.TCPReadBufferSize, lr.Config.TCPTimeout, remoteTimeout, recio.ModeManual)

	errored := false

	logWriter.HandleSync(func(progress log.SyncProgress) {

		// If an error occurred during copy we
		// wont try to send ack back to client.
		if errored {
			return
		}

		_, err = tr.WriteAck(&progress)
		if err != nil {
			logger.Debug(err)
			return
		}

		err = tr.Flush()
		if err != nil {
			logger.Debug(err)
			return
		}
	})

	err = writeTCP(logWriter, tr)
	if err != nil {

		errored = true
		logger.Debug(err)

		// Close log writer first to ensure sync handler wont
		// send sync progress anymore to the client.
		logWriter.Close()

		// Send error to the client to give it
		// a chance to close gracefully.
		tr.WriteError(err)
		tr.Flush()

		// Finaly close tcp conn.
		tr.Close()
		return
	}

	err = logWriter.Close()
	if err != nil {
		logger.Debug(err)

		tr.Close()
		return
	}

	err = tr.Close()
	if err != nil {
		logger.Debug(err)
	}
}

func writeTCP(lw *log.FaninWriter, tr *tcp.TCPReader) (err error) {

	record := log.Record{}

	for {
		_, err = tr.Read(&record)
		if err == io.EOF {
			break
		}

		if err == recio.ErrMustFill {

			err = lw.Flush()
			if err != nil {
				return err
			}

			err = tr.Fill()
			if err != nil {
				return err
			}

			continue
		}

		if err != nil {
			return err
		}

		_, err = lw.Write(&record)
		if err != nil {
			return err
		}
	}

	err = lw.Flush()
	if err != nil {
		return err
	}

	return nil
}
