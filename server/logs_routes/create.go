package logs_routes

import (
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
)

func (lr *LogsRouter) CreateHandler(w http.ResponseWriter, r *http.Request) {

	config := log.DefaultConfig

	form := api.CreateLogForm{
		Name:      "",
		LogConfig: (*api.LogConfig)(&config),
	}

	err := r.ParseForm()
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = lr.schemaDecoder.Decode(&form, r.PostForm)
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
		return
	}

	ml, err := lr.manager.CreateLog(form.Name, config)
	if err == log.ErrExist {
		api.WriteError(w, http.StatusBadRequest, api.ErrLogExist)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	logInfo := ml.Stat()

	api.WriteResponse(w, http.StatusOK, api.CreateLogResponse(logInfo))
}
