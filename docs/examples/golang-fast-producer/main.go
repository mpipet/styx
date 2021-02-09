package main

import (
	"encoding/json"
	logger "log"
	"time"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
)

type Event struct {
	Timestamp int64 `json:"timestamp"`
	Payload string `json:"payload"`
}

func main() {

	styxURL := "http://localhost:8000"
	logName := "fast"
	bufferSize := 1 << 20 // 1MB
	readTimeout := 30 // 30 seconds

	c := client.NewClient(styxURL)

	tw, err := c.WriteRecordsTCP(logName, recio.ModeAuto, bufferSize, readTimeout)
	if err != nil {
		logger.Fatal(err)
	}

	count := 0

	for {
		event := Event{
			Timestamp: time.Now().Unix(),
			Payload: "Hello, Styx !",
		}

		payload, err := json.Marshal(event)
		if err != nil {
			logger.Fatal(err)
		}

		r := log.Record(payload)

		_, err = tw.Write(&r)
		if err != nil {
			logger.Fatal(err)
		}

		err = tw.Flush()
		if err != nil {
			logger.Fatal(err)
		}

		count += 1

		if count % 1000 == 0 {
			logger.Printf("sent %d records", count)
		}
	}

	err = tw.Flush()
	if err != nil {
		logger.Fatal(err)
	}

	err = tw.Close()
	if err != nil {
		logger.Fatal(err)
	}

}
