package config

import (
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/metrics"
	"gitlab.com/dataptive/styx/metrics/statsd"

	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	PIDFile                string                `toml:"pid_file"`
	BindAddress            string                `toml:"bind_address"`
	ShutdownTimeout        int                   `toml:"shutdown_timeout"`
	CORSAllowedOrigins     []string              `toml:"cors_allowed_origins"`
	HTTPWriteBufferSize    int                   `toml:"http_write_buffer_size"`
	HTTPReadBufferSize     int                   `toml:"http_read_buffer_size"`
	HTTPFollowTimeout      int                   `toml:"http_follow_timeout"`
	HTTPMaxFollowTimeout   int                   `toml:"http_max_follow_timeout"`
	TCPReadBufferSize      int                   `toml:"tcp_read_buffer_size"`
	TCPWriteBufferSize     int                   `toml:"tcp_write_buffer_size"`
	TCPTimeout             int                   `toml:"tcp_timeout"`
	LogManager             TOMLLogManagerConfig  `toml:"log_manager"`
	Metrics                TOMLMetricsConfig     `toml:"metrics"`
}


type TOMLLogManagerConfig struct {
	DataDirectory   string `toml:"data_directory"`
	WriteBufferSize int    `toml:"write_buffer_size"`
}

type TOMLMetricsConfig struct {
	Statsd     *TOMLStatsdConfig `toml:"statsd"`
}

type TOMLStatsdConfig struct {
	Protocol      string `toml:"protocol"`
	Address       string `toml:"address"`
	Prefix        string `toml:"prefix"`
}

type Config struct {
	PIDFile                string
	BindAddress            string
	ShutdownTimeout        int
	CORSAllowedOrigins     []string
	HTTPReadBufferSize     int
	HTTPWriteBufferSize    int
	HTTPFollowTimeout      int
	HTTPMaxFollowTimeout   int
	TCPReadBufferSize      int
	TCPWriteBufferSize     int
	TCPTimeout             int
	LogManager             logman.Config
	Metrics                metrics.Config
}

func Load(path string) (c Config, err error) {

	tc := &TOMLConfig{}

	_, err = toml.DecodeFile(path, tc)
	if err != nil {
		return c, err
	}

	c.PIDFile = tc.PIDFile
	c.BindAddress = tc.BindAddress
	c.ShutdownTimeout = tc.ShutdownTimeout
	c.CORSAllowedOrigins = tc.CORSAllowedOrigins
	c.HTTPReadBufferSize = tc.HTTPReadBufferSize
	c.HTTPWriteBufferSize = tc.HTTPWriteBufferSize
	c.HTTPFollowTimeout = tc.HTTPFollowTimeout
	c.HTTPMaxFollowTimeout = tc.HTTPMaxFollowTimeout
	c.TCPReadBufferSize = tc.TCPReadBufferSize
	c.TCPWriteBufferSize = tc.TCPWriteBufferSize
	c.TCPTimeout = tc.TCPTimeout
	c.LogManager = logman.Config(tc.LogManager)
	c.Metrics = metrics.Config{
		Statsd: (*statsd.Config)(tc.Metrics.Statsd),
	}

	return c, nil
}
