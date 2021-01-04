package config

import (
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/metrics"
	"gitlab.com/dataptive/styx/metrics/statsd"
	"gitlab.com/dataptive/styx/nodeman"

	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	PIDFile                string                `toml:"pid_file"`
	BindAddress            string                `toml:"bind_address"`
	ShutdownTimeout        int                   `toml:"shutdown_timeout"`
	CORSAllowedOrigins     []string              `toml:"cors_allowed_origins"`
	HTTPWriteBufferSize    int                   `toml:"http_write_buffer_size"`
	HTTPReadBufferSize     int                   `toml:"http_read_buffer_size"`
	HTTPLongpollTimeout    int                   `toml:"http_longpoll_timeout"`
	HTTPMaxLongpollTimeout int                   `toml:"http_max_longpoll_timeout"`
	TCPReadBufferSize      int                   `toml:"tcp_read_buffer_size"`
	TCPWriteBufferSize     int                   `toml:"tcp_write_buffer_size"`
	TCPTimeout             int                   `toml:"tcp_timeout"`
	LogManager             TOMLLogManagerConfig  `toml:"log_manager"`
	NodeManager            TOMLNodeManagerConfig `toml:"node_manager"`
	Metrics                TOMLMetricsConfig     `toml:"metrics"`
}


type TOMLLogManagerConfig struct {
	DataDirectory   string `toml:"data_directory"`
	WriteBufferSize int    `toml:"write_buffer_size"`
}

type TOMLNodeManagerConfig struct {
	NodeName         string `toml:"node_name"`
	RaftDirectory    string `toml:"raft_directory"`
	AdvertiseAddress string `toml:"advertise_address"`
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
	HTTPLongpollTimeout    int
	HTTPMaxLongpollTimeout int
	TCPReadBufferSize      int
	TCPWriteBufferSize     int
	TCPTimeout             int
	LogManager             logman.Config
	NodeManager            nodeman.Config
	Metrics                metrics.Config
}

func Load(path string) (c Config, err error) {

	tc := &TOMLConfig{}

	_, err = toml.DecodeFile(path, tc)
	if err != nil {
		return c, err
	}

	if tc.NodeManager.AdvertiseAddress == "" {
		tc.NodeManager.AdvertiseAddress = tc.BindAddress
	}

	c.PIDFile = tc.PIDFile
	c.BindAddress = tc.BindAddress
	c.ShutdownTimeout = tc.ShutdownTimeout
	c.HTTPReadBufferSize = tc.HTTPReadBufferSize
	c.HTTPWriteBufferSize = tc.HTTPWriteBufferSize
	c.HTTPLongpollTimeout = tc.HTTPLongpollTimeout
	c.HTTPMaxLongpollTimeout = tc.HTTPMaxLongpollTimeout
	c.TCPReadBufferSize = tc.TCPReadBufferSize
	c.TCPWriteBufferSize = tc.TCPWriteBufferSize
	c.TCPTimeout = tc.TCPTimeout
	c.LogManager = logman.Config(tc.LogManager)
	c.NodeManager = nodeman.Config(tc.NodeManager)
	c.Metrics = metrics.Config{
		Statsd: (*statsd.Config)(tc.Metrics.Statsd),
	}

	return c, nil
}
