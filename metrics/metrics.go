package metrics

import (
	"gitlab.com/dataptive/styx/log"

	"gitlab.com/dataptive/styx/metrics/prometheus"
	"gitlab.com/dataptive/styx/metrics/statsd"
)

type Reporter interface {
	ReportLogStats(string, log.Stat) error

	Close() error
}

type MetricsReporter struct {
	reporters []Reporter
}

func NewMetricsReporter(config Config) (mp *MetricsReporter, err error) {

	var reporters []Reporter

	pp := prometheus.NewPrometheusReporter()
	reporters = append(reporters, pp)

	if config.Statsd != nil {
		sp, err := statsd.NewStatsdReporter(*config.Statsd)
		if err != nil {
			return nil, err
		}

		reporters = append(reporters, sp)
	}

	mp = &MetricsReporter{
		reporters: reporters,
	}

	return mp, nil
}

func (mp *MetricsReporter) ReportLogStats(name string, stats log.Stat) (err error) {

	for _, reporter := range mp.reporters {
		reporter.ReportLogStats(name, stats)
	}

	return nil
}

func (mp *MetricsReporter) Close() (err error) {

	for _, reporter := range mp.reporters {
		reporter.Close()
	}

	return nil
}
