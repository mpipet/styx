package logs_routes

import (
	"io"
	"net/http"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/api/tcp"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) ReadTCPHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	vars := mux.Vars(r)
	name := vars["name"]

	remoteTimeout := lr.config.TCPTimeout

	// TODO: Change the header name to a more adequate one.
	rawTimeout := r.Header.Get(api.TimeoutHeaderName)
	if rawTimeout != "" {

		remoteTimeout, err = strconv.Atoi(rawTimeout)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
			logger.Debug(err)
			return
		}
	}

	params := api.ReadRecordsTCPParams{
		Whence:   log.SeekOrigin,
		Position: 0,
		Count:    10,
		Follow: false,
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

	logReader, err := managedLog.NewReader(lr.config.HTTPReadBufferSize, params.Follow, recio.ModeManual)
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

	w.Header().Add(api.TimeoutHeaderName, strconv.Itoa(lr.config.TCPTimeout))
	conn, err := UpgradeTCP(w)
	if err != nil {
		logger.Debug(err)
		logReader.Close()
		return
	}

	err = conn.SetReadBuffer(lr.config.TCPReadBufferSize)
	if err != nil {
		logger.Warn(err)
	}

	err = conn.SetWriteBuffer(lr.config.TCPWriteBufferSize)
	if err != nil {
		logger.Warn(err)
	}

	tcpWriter := tcp.NewTCPWriter(conn, lr.config.TCPWriteBufferSize, lr.config.TCPReadBufferSize, lr.config.TCPTimeout, remoteTimeout, recio.ModeAuto)

	tcpWriter.HandleError(func(err error) {
		logger.Debug(err)

		// Close reader to unlock follow.
		logReader.Close()
	})

	err = readTCP(tcpWriter, logReader, params.Count, params.Follow)
	if err != nil {
		logger.Debug(err)

		// Close reader to unlock follow
		// if not already done.
		logReader.Close()

		// Try to write error back to
		// client in case conn is still open.
		tcpWriter.WriteError(err)
		tcpWriter.Flush()

		// Close conn in case its still open.
		tcpWriter.Close()

		return
	}

	err = logReader.Close()
	if err != nil {
		logger.Debug(err)

		tcpWriter.Close()
	}

	err = tcpWriter.Close()
	if err != nil {
		logger.Debug(err)
	}
}

func readTCP(w *tcp.TCPWriter, lr *log.LogReader, limit int64, follow bool) (err error) {

	count := int64(0)
	record := log.Record{}

	for {
		if count == limit {
			break
		}

		_, err := lr.Read(&record)
		if err == io.EOF {

			if follow == false {
				break
			}

			continue
		}

		if err == recio.ErrMustFill {

			err = w.Flush()
			if err != nil {
				return err
			}

			err = lr.Fill()
			if err != nil {
				return err
			}

			continue
		}

		if err != nil {
			return err
		}

		_, err = w.Write(&record)
		if err != nil {
			return err
		}

		count++
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}
