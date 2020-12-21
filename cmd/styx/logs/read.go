package logs

import (
	"io"
	"os"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
	"gitlab.com/dataptive/styx/recio/recioutil"
	// "gitlab.com/dataptive/styx/util"

	"github.com/spf13/pflag"
)

const logsReadUsage = `
Usage: styx logs read NAME [OPTIONS]

Read from log and output line delimited record payloads

Options:
	--position int 		Position to start reading from (default 0)
	--whence string		Reference from which position is computed [origin|start|end] (default "start")
	--count int		Maximum count of records to read (cannot be used in association with --follow)
	--follow 		Wait for new records when reaching end of stream
	--unbuffered		Do not buffer read
	--binary		Output binary records

Global Options:
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

func ReadLog(args []string) {

	const (
		readBufferSize = 1 << 20 // 1MB
		writeBufferSize = 1 << 20 // 1MB
		timeout = 100
	)

	readOpts := pflag.NewFlagSet("read", pflag.ContinueOnError)
	whence := readOpts.String("whence", string(log.SeekOrigin), "")
	position := readOpts.Int64("position", 0, "")
	count := readOpts.Int64("count", -1, "")
	follow := readOpts.Bool("follow", false, "")
	unbuffered := readOpts.Bool("unbuffered", false, "")
	binary := readOpts.Bool("binary", false, "")
	host := readOpts.String("host", "http://localhost:8000", "")
	isHelp := readOpts.Bool("help", false, "")
	readOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	err := readOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsReadUsage)
	}

	if readOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	httpClient := client.NewClient(*host)

	params := api.ReadRecordsTCPParams{
		Whence: log.Whence(*whence),
		Position: *position,
		Count: *count,
		Follow: *follow,
	}

	tcpReader, err := httpClient.ReadRecordsTCP(readOpts.Args()[0], params, recio.ModeAuto, readBufferSize, timeout)
	if err != nil {
		cmd.DisplayError(err)
	}

	bufferedWriter := recio.NewBufferedWriter(os.Stdout, writeBufferSize, recio.ModeAuto)

	isTerm, err := cmd.IsTerminal(os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}

	mustFlush := isTerm || *unbuffered

	var encoder recio.Encoder
	record := &log.Record{}
	for {
		_, err := tcpReader.Read(record)
		if err == io.EOF {
			break
		}

		if err != nil {
			cmd.DisplayError(err)
		}


		if *binary {
			encoder = record
		} else {
			encoder = (*recioutil.Line)(record)
			// line, isLine := (*log.Record)(decoder.(*recioutil.Line))
			// record = (*log.Record)(decoder.(*recioutil.Line))
			// record = (*log.Record)(line)
		}

		_, err = bufferedWriter.Write(encoder)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = bufferedWriter.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}
	}

	err = bufferedWriter.Flush()
	if err != nil {
		cmd.DisplayError(err)
	}

	err = tcpReader.Close()
	if err != nil {
		cmd.DisplayError(err)
	}
}
