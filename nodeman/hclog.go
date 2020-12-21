package nodeman

import (
	"fmt"

	"gitlab.com/dataptive/styx/logger"

	"github.com/hashicorp/go-hclog"
)

func newHCLogger() (hcLogger hclog.Logger) {

	interceptLogger := hclog.NewInterceptLogger(&hclog.LoggerOptions{})

	sinkAdapter := &sinkAdapter{}
	interceptLogger.RegisterSink(sinkAdapter)
	interceptLogger.SetLevel(hclog.Off)

	return interceptLogger
}

type sinkAdapter struct{}

func (sa *sinkAdapter) Accept(name string, level hclog.Level, msg string, args ...interface{}) {

	argsStr := dumpArgs(args)

	switch level {
	case hclog.Trace:
		logger.Trace("nodeman:", msg, argsStr)
	case hclog.Debug:
		logger.Debug("nodeman:", msg, argsStr)
	case hclog.Info:
		logger.Info("nodeman:", msg, argsStr)
	case hclog.Warn:
		logger.Warn("nodeman:", msg, argsStr)
	case hclog.Error:
		logger.Error("nodeman:", msg, argsStr)
	default:
		//
	}
}

func dumpArgs(args []interface{}) string {

	if len(args) == 0 {
		return ""
	}

	str := "("

	for i, v := range args {
		str += fmt.Sprintf("%v", v)

		if i == len(args)-1 {
			break
		}

		if i%2 == 0 {
			str += "="
		} else {
			str += ", "
		}
	}

	str += ")"

	return str
}
