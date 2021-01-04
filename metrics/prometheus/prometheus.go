package prometheus

import (
	"gitlab.com/dataptive/styx/log"

	prom "github.com/prometheus/client_golang/prometheus"
)

type PrometheusReporter struct {
	logRecordCount *prom.GaugeVec
	logFileSize    *prom.GaugeVec
}

func NewPrometheusReporter() (pp *PrometheusReporter) {

	logRecordCount := prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "log_record_count",
			Help: "Current record count",
		},
		[]string{"log"},
	)

	logFileSize := prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "log_file_size",
			Help: "Current log file size",
		},
		[]string{"log"},
	)

	prom.MustRegister(logRecordCount)
	prom.MustRegister(logFileSize)

	pp = &PrometheusReporter{
		logRecordCount: logRecordCount,
		logFileSize:    logFileSize,
	}

	return pp
}

func (pp *PrometheusReporter) Close() (err error) {

	return nil
}

func (pp *PrometheusReporter) ReportLogStats(name string, stats log.Stat) (err error) {

	recordCount := float64(stats.EndPosition - stats.StartPosition)
	pp.logRecordCount.
		With(prom.Labels{"log": name}).
		Set(recordCount)

	fileSize := float64(stats.EndOffset - stats.StartOffset)
	pp.logFileSize.
		With(prom.Labels{"log": name}).
		Set(fileSize)

	return nil
}
