package logs

import (
	"os"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsRestoreUsage = `
Usage: styx logs restore NAME [OPTIONS]

Restore log

Global Options:
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

func RestoreLog(args []string) {
	restoreOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := restoreOpts.String("host", "http://localhost:8000", "")
	isHelp := restoreOpts.Bool("help", false, "")
	restoreOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	err := restoreOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsRestoreUsage)
	}

	httpClient := client.NewClient(*host)

	if restoreOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	err = httpClient.RestoreLog(args[0], os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}
}
