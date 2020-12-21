package logs

import (
	"os"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsBackupUsage = `
Usage: styx logs backup NAME [OPTIONS]

Backup log

Global Options:
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

func BackupLog(args []string) {

	backupOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := backupOpts.String("host", "http://localhost:8000", "")
	isHelp := backupOpts.Bool("help", false, "")
	backupOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	err := backupOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsBackupUsage)
	}

	httpClient := client.NewClient(*host)

	if backupOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	err = httpClient.BackupLog(args[0], os.Stdout)
	if err != nil {
		cmd.DisplayError(err)
	}
}