package main

import (
	"os"

	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/cmd/styx/logs"
)

const (
	cliUsage = `
Usage: styx COMMAND

A command line interface (CLI) for the Styx API.

Commands:
	logs  Manage logs

Global Options:
	--format string			Output format [text|json] (default "text")
	--host string 			Server to connect to (default "http://localhost:8000")
	--help 				Display help
`

	logsUsage = `
Usage: styx logs COMMAND

Manage logs

Commands:
	list		List available logs
	create		Create a new log
	get		Show log details
	delete		Delete a log
	backup		Backup a log
	restore		Restore a log
	write		Write records to a log
	read		Read records from a log

Global Options:
	--format string			Output format [text|json] (default "text")
	--host string 			Server to connect to (default "http://localhost:8000")
	--help 				Display help
`
)

func main() {

	args := os.Args[1:]

	if len(args) < 1 {
		cmd.DisplayUsage(cmd.MisuseCode, cliUsage)
	}

	switch args[0] {
	case "logs":

		if len(args) < 2 {
			cmd.DisplayUsage(cmd.MisuseCode, logsUsage)
		}

		args = args[1:]

		switch args[0] {
		case "list":
			logs.ListLogs(args[1:])
		case "create":
			logs.CreateLog(args[1:])
		case "get":
			logs.GetLog(args[1:])
		case "delete":
			logs.DeleteLog(args[1:])
		case "backup":
			logs.BackupLog(args[1:])
		case "restore":
			logs.RestoreLog(args[1:])
		case "write":
			logs.WriteLog(args[1:])
		case "read":
			logs.ReadLog(args[1:])
		case "--help":
			cmd.DisplayUsage(cmd.SuccessCode, logsUsage)
		default:
			cmd.DisplayUsage(cmd.MisuseCode, logsUsage)
		}

	case "--help":
		cmd.DisplayUsage(cmd.SuccessCode, cliUsage)
	default:
		cmd.DisplayUsage(cmd.MisuseCode, cliUsage)
	}
}
