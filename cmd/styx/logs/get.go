package logs

import (
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsGetUsage = `
Usage: styx logs get NAME [OPTIONS]

Show log details

Global Options:
	--format string		Output format [text|json] (default "text")
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

const logsGetTmpl = `name:	{{.Name}}
status:	{{.Status}}
record_count:	{{.RecordCount}}
file_size:	{{.FileSize}}
start_position:	{{.StartPosition}}
end_position:	{{.EndPosition}}
`

func GetLog(args []string) {

	getOpts := pflag.NewFlagSet("logs get", pflag.ContinueOnError)
	host := getOpts.String("host", "http://localhost:8000", "")
	format := getOpts.String("format", "text", "")
	isHelp := getOpts.Bool("help", false, "")
	getOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	err := getOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsGetUsage)
	}

	httpClient := client.NewClient(*host)

	if getOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	log, err := httpClient.GetLog(args[0])

	if err != nil {
		cmd.DisplayError(err)
	}

	if *format == "json" {
		cmd.DisplayAsJSON(log)
		return
	}

	cmd.DisplayAsDefault(logsGetTmpl, log)
}
