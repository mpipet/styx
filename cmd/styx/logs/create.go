package logs

import (
	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/log"

	"github.com/spf13/pflag"
)

const logsCreateUsage = `
Usage: styx logs create NAME [OPTIONS]

Create a new log

Options:
	--max-record-size bytes 	Maximum record size
	--index-after-size bytes 	Write a segment index entry after every size
	--segment-max-count records	Create a new segment when current segment exceeds this number of records
	--segment-max-size bytes	Create a new segment when current segment exceeds this size
	--segment-max-age seconds	Create a new segment when current segment exceeds this age
	--log-max-count records 	Expire oldest segment when log exceeds this number of records
	--log-max-size bytes 		Expire oldest segment when log exceeds this size
	--log-max-age seconds 		Expire oldest segment when log exceeds this age

Global Options:
	--format string			Output format [text|json] (default "text")
	--host string 			Server to connect to (default "http://localhost:8000")
	--help 				Display help
`

const logsCreateTmpl = `name:	{{.Name}}
status:	{{.Status}}
record_count:	{{.RecordCount}}
file_size:	{{.FileSize}}
start_position:	{{.StartPosition}}
end_position:	{{.EndPosition}}
`

func CreateLog(args []string) {

	createOpts := pflag.NewFlagSet("logs create", pflag.ContinueOnError)
	maxRecordSize := createOpts.Int("max-record-size", log.DefaultConfig.MaxRecordSize, "")
	indexAfterSize := createOpts.Int64("index-after-size", log.DefaultConfig.IndexAfterSize, "")
	segmentMaxCount := createOpts.Int64("segment-max-count", log.DefaultConfig.SegmentMaxCount, "")
	segmentMaxSize := createOpts.Int64("segment-max-size", log.DefaultConfig.SegmentMaxSize, "")
	segmentMaxAge := createOpts.Int64("segment-max-age", log.DefaultConfig.SegmentMaxAge, "")
	logMaxCount := createOpts.Int64("log-max-count", log.DefaultConfig.LogMaxCount, "")
	logMaxSize := createOpts.Int64("log-max-size", log.DefaultConfig.LogMaxSize, "")
	logMaxAge := createOpts.Int64("log-max-age", log.DefaultConfig.LogMaxAge, "")
	format := createOpts.String("format", "text", "")
	host := createOpts.String("host", "http://localhost:8000", "")
	isHelp := createOpts.Bool("help", false, "")
	createOpts.Usage = func() {
		cmd.DisplayUsage(2, logsCreateUsage)
	}

	err := createOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(2, logsCreateUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(0, logsCreateUsage)
	}

	if createOpts.NArg() != 1 {
		cmd.DisplayUsage(2, logsCreateUsage)
	}

	httpClient := client.NewClient(*host)

	name := createOpts.Args()[0]
	config := api.LogConfig{
		MaxRecordSize: *maxRecordSize,
		IndexAfterSize:  *indexAfterSize,
		SegmentMaxCount: *segmentMaxCount,
		SegmentMaxSize:  *segmentMaxSize,
		SegmentMaxAge: *segmentMaxAge,
		LogMaxCount: *logMaxCount,
		LogMaxSize: *logMaxSize,
		LogMaxAge: *logMaxAge,
	}

	log, err := httpClient.CreateLog(name, config)
	if err != nil {
		cmd.DisplayError(err)
	}

	if *format == "json" {
		cmd.DisplayAsJSON(log)
		return
	}

	cmd.DisplayAsDefault(logsCreateTmpl, log)
}
