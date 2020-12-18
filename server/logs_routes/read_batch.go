package logs_routes

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/clock"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/manager"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

var (
	now = clock.New(time.Second)
)

func (lr *LogsRouter) ReadRecordsBatchHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	params := api.ReadRecordsBatchParams{
		Whence:   log.SeekOrigin,
		Position: 0,
		Count:    100,
		Longpoll: false,
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

	timeout := lr.Config.HTTPLongpollTimeout

	if params.Longpoll {

		rawTimeout := r.Header.Get("Request-Timeout")
		if rawTimeout != "" {

			timeout, err = strconv.Atoi(rawTimeout)
			if err != nil {
				api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
				logger.Debug(err)
				return
			}
		}

		// Limit the timeout as defined in config.
		if timeout > lr.Config.HTTPMaxLongpollTimeout {
			timeout = lr.Config.HTTPMaxLongpollTimeout
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

	bufferedWriter := recio.NewBufferedWriter(w, lr.Config.HTTPWriteBufferSize, recio.ModeAuto)

	logReader, err := managedLog.NewReader(lr.Config.HTTPReadBufferSize, params.Longpoll, recio.ModeManual)
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	fillTimeout := time.Duration(timeout) * time.Second

	err = readBatch(bufferedWriter, logReader, params.Count, params.Longpoll, fillTimeout)
	if err != nil {
		logger.Debug(err)
		logReader.Close()
		return
	}

	err = logReader.Close()
	if err != nil {
		logger.Debug(err)
	}
}

func readBatch(bw *recio.BufferedWriter, lr *log.LogReader, limit int64, longPoll bool, timeout time.Duration) (err error) {

	count := int64(0)
	record := log.Record([]byte{})

	for {
		if count == limit {
			break
		}

		_, err := lr.Read(&record)
		if err == io.EOF {
			break
		}

		if err == recio.ErrMustFill {

			err = bw.Flush()
			if err != nil {
				return err
			}

			if longPoll {
				start := time.Now()
				deadline := start.Add(timeout)

				lr.SetWaitDeadline(deadline)
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

		_, err = bw.Write(&record)
		if err != nil {
			return err
		}

		count++
	}

	err = bw.Flush()
	if err != nil {
		return err
	}

	return nil
}
