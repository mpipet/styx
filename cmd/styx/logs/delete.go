package logs

import (
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsDeleteUsage = `
Usage: styx logs delete NAME [OPTIONS]

Delete a log

Global Options:
	--host string 	Server to connect to (default "http://localhost:8000")
	--help 		Display help
`

func DeleteLog(args []string) {

	deleteOpts := pflag.NewFlagSet("logs delete", pflag.ContinueOnError)
	host := deleteOpts.String("host", "http://localhost:8000", "")
	isHelp := deleteOpts.Bool("help", false, "")
	deleteOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	err := deleteOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsDeleteUsage)
	}

	httpClient := client.NewClient(*host)

	if deleteOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	err = httpClient.DeleteLog(deleteOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}
}
