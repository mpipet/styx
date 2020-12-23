package logs

import (
	"fmt"
	"time"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsListUsage = `
Usage: styx logs list [OPTIONS]

List available logs

Global Options:
	--watch			Display and update informations about logs
	--format string		Output format [text|json] (default "text")
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

const logsListTmpl = `NAME	STATUS	RECORD COUNT	FILE SIZE	START POSITION	END POSITION
{{range .}}{{.Name}}	{{.Status}}	{{.RecordCount}}	{{.FileSize}}	{{.StartPosition}}	{{.EndPosition}}
{{end}}`

func ListLogs(args []string) {

	listOpts := pflag.NewFlagSet("logs list", pflag.ContinueOnError)
	watch := listOpts.Bool("watch", false, "")
	format := listOpts.String("format", "default", "")
	host := listOpts.String("host", "http://localhost:8000", "")
	isHelp := listOpts.Bool("help", false, "")
	listOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	err := listOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsListUsage)
	}

	httpClient := client.NewClient(*host)

	if listOpts.NArg() != 0 {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	for {
		logs, err := httpClient.ListLogs()
		if err != nil {
			cmd.DisplayError(err)
		}

		if *watch {
			// Clear terminal
			fmt.Printf("\033[H\033[2J")
		}

		if *format == "json" {
			cmd.DisplayAsJSON(logs)

		} else {
			cmd.DisplayAsDefault(logsListTmpl, logs)
		}

		if *watch {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}
