package config

import (
	"gitlab.com/dataptive/styx/manager"

	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	PIDFile            string            `toml:"pid_file"`
	BindAddress        string            `toml:"bind_address"`
	ShutdownTimeout    int               `toml:"shutdown_timeout"`
	CORSAllowedOrigins []string          `toml:"cors_allowed_origins"`
	HTTPReadBufferSize int               `toml:"http_read_buffer_size"`
	LogManager         TOMLManagerConfig `toml:"logs"`
}

type TOMLManagerConfig struct {
	DataDirectory   string `toml:"data_directory"`
	WriteBufferSize int    `toml:"write_buffer_size"`
}

type Config struct {
	PIDFile            string
	BindAddress        string
	ShutdownTimeout    int
	CORSAllowedOrigins []string
	HTTPReadBufferSize int
	LogManager         manager.Config
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
	c.LogManager = manager.Config(tc.LogManager)

	return c, nil
}
