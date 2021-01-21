package logs_routes

import (
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"
	"gitlab.com/dataptive/styx/recio/recioutil"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) ReadLinesMatcher(r *http.Request, rm *mux.RouteMatch) (match bool) {

	accept := r.Header.Get("Accept")
	mediaType, _, _ := mime.ParseMediaType(accept)

	match = mediaType == api.RecordLinesMediaType

	return match
}

func (lr *LogsRouter) ReadLinesHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	accept := r.Header.Get("Accept")
	_, typeParams, err := mime.ParseMediaType(accept)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	if typeParams["line-ending"] == "" {
		typeParams["line-ending"] = "lf"
	}

	delimiter, valid := recioutil.LineEndings[typeParams["line-ending"]]
	if !valid {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	params := api.ReadRecordsLinesParams{
		Whence:   log.SeekOrigin,
		Position: 0,
		Count:    -1,
		Follow:   false,
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

	timeout := lr.config.HTTPFollowTimeout

	if params.Follow {

		rawTimeout := r.Header.Get(api.TimeoutHeaderName)
		if rawTimeout != "" {

			timeout, err = strconv.Atoi(rawTimeout)
			if err != nil {
				api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
				logger.Debug(err)
				return
			}
		}

		// Limit the timeout as defined in config.
		if timeout > lr.config.HTTPMaxFollowTimeout {
			timeout = lr.config.HTTPMaxFollowTimeout
		}
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

	bufferedWriter := recio.NewBufferedWriter(w, lr.config.HTTPWriteBufferSize, recio.ModeAuto)
	lineWriter := recioutil.NewLineWriter(bufferedWriter, delimiter)

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

	mediaType := mime.FormatMediaType(api.RecordLinesMediaType, typeParams)
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(http.StatusOK)

	fillTimeout := time.Duration(timeout) * time.Second

	err = readLines(lineWriter, bufferedWriter, logReader, params.Count, params.Follow, fillTimeout)
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

func readLines(lw *recioutil.LineWriter, bw *recio.BufferedWriter, lr *log.LogReader, limit int64, follow bool, timeout time.Duration) (err error) {

	count := int64(0)
	record := &log.Record{}

	for {
		if count == limit {
			break
		}

		_, err := lr.Read(record)
		if err == io.EOF {
			break
		}

		if err == recio.ErrMustFill {

			err = bw.Flush()
			if err != nil {
				return err
			}

			if follow {

				if count > 0 {
					break
				}

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

		_, err = lw.Write((*recioutil.Line)(record))
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
