package metrics

import (
	"gitlab.com/dataptive/styx/metrics/statsd"
)

var (
	DefaultConfig = Config{
		Statsd: &statsd.DefaultConfig,
	}
)

type Config struct {
	Statsd *statsd.Config
}
