package logs

import (
	"errors"
	"io"
	"os"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
	"gitlab.com/dataptive/styx/recio/recioutil"

	"github.com/spf13/pflag"
)

const logsWriteUsage = `
Usage: styx logs write NAME [OPTIONS]

Write to log, input is expected to be line delimited record payloads

Options:
	--unbuffered	Do not buffer writes
	--binary	Process input as binary records
	--line-ending   Line end [cr|lf|crlf] for non binary record output

Global Options:
	--host string 	Server to connect to (default "http://localhost:8000")
	--help 		Display help
`

func WriteLog(args []string) {

	const (
		readBufferSize = 1 << 20 // 1MB
		writeBufferSize = 1 << 20 // 1MB
		timeout = 100
	)

	writeOpts := pflag.NewFlagSet("logs write", pflag.ContinueOnError)
	unbuffered := writeOpts.Bool("unbuffered", false, "")
	binary := writeOpts.Bool("binary", false, "")
	lineEnding := writeOpts.String("line-ending", "lf", "")
	host := writeOpts.String("host", "http://localhost:8000", "")
	isHelp := writeOpts.Bool("help", false, "")
	writeOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	err := writeOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsWriteUsage)
	}

	httpClient := client.NewClient(*host)

	if writeOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	tcpWriter, err := httpClient.WriteRecordsTCP(writeOpts.Args()[0], recio.ModeAuto, writeBufferSize, timeout)
	if err != nil {
		cmd.DisplayError(err)
	}

	reader := recio.NewBufferedReader(os.Stdin, readBufferSize, recio.ModeAuto)

	var decoder recio.Decoder
	if *binary {
		decoder = &log.Record{}
	} else {
		switch(*lineEnding) {
		case "cr":
			decoder = &recioutil.LineCR{}
		case "lf":
			decoder = &recioutil.LineLF{}
		case "crlf":
			decoder = &recioutil.LineCRLF{}
		default:
			cmd.DisplayError(errors.New("unknown line ending"))
		}
	}

	isTerm, err := cmd.IsTerminal(os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}

	mustFlush := isTerm || *unbuffered

	record := &log.Record{}
	for {
		_, err := reader.Read(decoder)
		if err == io.EOF {
			break
		}

		if err != nil {
			cmd.DisplayError(err)
		}

		if *binary {
			// Convert back record as decoder to record
			record = decoder.(*log.Record)
		} else {
			// Convert line as decoder interface to record
			switch(*lineEnding) {
			case "cr":
				record = (*log.Record)(decoder.(*recioutil.LineCR))
			case "lf":
				record = (*log.Record)(decoder.(*recioutil.LineLF))
			case "crlf":
				record = (*log.Record)(decoder.(*recioutil.LineCRLF))
			}
		}

		_, err = tcpWriter.Write(record)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = tcpWriter.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}
	}

	err = tcpWriter.Flush()
	if err != nil {
		cmd.DisplayError(err)
	}

	err = tcpWriter.Close()
	if err != nil {
		cmd.DisplayError(err)
	}
}
