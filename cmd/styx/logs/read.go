// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logs

import (
	"errors"
	"io"
	"os"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
	"gitlab.com/dataptive/styx/recio/recioutil"

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
	--line-ending   	Line end [cr|lf|crlf] for non binary record output

Global Options:
	--host string 		Server to connect to (default "http://localhost:8000")
	--help 			Display help
`

func ReadLog(args []string) {

	const (
		readBufferSize  = 1 << 20 // 1MB
		writeBufferSize = 1 << 20 // 1MB
		timeout         = 100
	)

	readOpts := pflag.NewFlagSet("read", pflag.ContinueOnError)
	whence := readOpts.String("whence", string(log.SeekOrigin), "")
	position := readOpts.Int64("position", 0, "")
	count := readOpts.Int64("count", -1, "")
	follow := readOpts.Bool("follow", false, "")
	unbuffered := readOpts.Bool("unbuffered", false, "")
	binary := readOpts.Bool("binary", false, "")
	lineEnding := readOpts.String("line-ending", "lf", "")
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
		Whence:   log.Whence(*whence),
		Position: *position,
	}

	logInfo, err := httpClient.GetLog(readOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}

	if !*follow && *count == -1 {
		count = &logInfo.RecordCount
	}

	tcpReader, err := httpClient.ReadRecordsTCP(readOpts.Args()[0], params, recio.ModeAuto, readBufferSize, timeout)
	if err != nil {
		cmd.DisplayError(err)
	}

	var writer recio.Writer
	var encoder recio.Encoder

	bufferedWriter := recio.NewBufferedWriter(os.Stdout, writeBufferSize, recio.ModeAuto)
	writer = bufferedWriter

	if !*binary {
		var delimiter []byte
		encoder = &recioutil.Line{}

		delimiter, valid := recioutil.LineEndings[*lineEnding]
		if !valid {
			cmd.DisplayError(errors.New("unknown line ending"))
		}

		writer = recioutil.NewLineWriter(bufferedWriter, delimiter)
	}

	isTerm, err := cmd.IsTerminal(os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}

	mustFlush := isTerm || *unbuffered
	record := &log.Record{}
	read := int64(0)
	for {
		if !*follow && read == *count {
			break
		}

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
		}

		_, err = writer.Write(encoder)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = bufferedWriter.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}

		read++
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
