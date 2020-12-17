package main

import (
	"fmt"
	"os"
	"strings"

	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/server"
	"gitlab.com/dataptive/styx/server/config"

	"github.com/spf13/pflag"
)

const (
	defaultConfigPath = "config.toml"
)

const usage = `
Usage: styx-server [OPTIONS]

Run Styx server

Options:
	--config string 	Config file path
	--log-level string 	Set the logging level [TRACE|DEBUG|INFO|WARN|ERROR|FATAL] (default "INFO")
	--help			Display help
`

func main() {

	options := pflag.NewFlagSet("", pflag.ContinueOnError)
	configPath := options.String("config", defaultConfigPath, "")
	level := options.String("log-level", "INFO", "")
	help := options.Bool("help", false, "")

	err := options.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(usage))
		os.Exit(2)
	}

	if options.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(usage))
		os.Exit(2)
	}

	if *help {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(usage))
		os.Exit(0)
	}

	logsLevels := map[string]int{
		"TRACE": logger.LevelTrace,
		"DEBUG": logger.LevelDebug,
		"INFO": logger.LevelInfo,
		"WARN": logger.LevelWarn,
		"ERROR": logger.LevelError,
		"FATAL": logger.LevelFatal,
	}

	logLevel, exists := logsLevels[*level]
	if !exists {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(usage))
		os.Exit(2)
	}

	logger.SetLevel(logLevel)

	serverConfig, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal(err)
	}

	styxServer, err := server.NewServer(serverConfig)
	if err != nil {
		logger.Fatal(err)
	}

	err = styxServer.Run()
	if err != nil {
		logger.Fatal(err)
	}
}
