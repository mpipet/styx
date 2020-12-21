package config

import (
	"gitlab.com/dataptive/styx/logman"

	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	PIDFile                string            `toml:"pid_file"`
	BindAddress            string            `toml:"bind_address"`
	ShutdownTimeout        int               `toml:"shutdown_timeout"`
	CORSAllowedOrigins     []string          `toml:"cors_allowed_origins"`
	HTTPWriteBufferSize    int               `toml:"http_write_buffer_size"`
	HTTPReadBufferSize     int               `toml:"http_read_buffer_size"`
	HTTPLongpollTimeout    int               `toml:"http_longpoll_timeout"`
	HTTPMaxLongpollTimeout int               `toml:"http_max_longpoll_timeout"`
	TCPReadBufferSize      int               `toml:"tcp_read_buffer_size"`
	TCPWriteBufferSize     int               `toml:"tcp_write_buffer_size"`
	TCPTimeout             int               `toml:"tcp_timeout"`
	LogManager             TOMLManagerConfig `toml:"logs"`
}

type TOMLManagerConfig struct {
	DataDirectory   string `toml:"data_directory"`
	WriteBufferSize int    `toml:"write_buffer_size"`
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
	c.HTTPReadBufferSize = tc.HTTPReadBufferSize
	c.HTTPWriteBufferSize = tc.HTTPWriteBufferSize
	c.HTTPLongpollTimeout = tc.HTTPLongpollTimeout
	c.HTTPMaxLongpollTimeout = tc.HTTPMaxLongpollTimeout
	c.TCPReadBufferSize = tc.TCPReadBufferSize
	c.TCPWriteBufferSize = tc.TCPWriteBufferSize
	c.TCPTimeout = tc.TCPTimeout
	c.LogManager = logman.Config(tc.LogManager)

	return c, nil
}
