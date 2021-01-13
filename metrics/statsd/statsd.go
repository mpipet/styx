package statsd

import (
	"fmt"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
)

const (
	recordCountPattern = "log.%s.record.count"
	fileSizePattern    = "log.%s.file.size"
)

type StatsdReporter struct {
	conn   *UDPConn
	client *Client
}

func NewStatsdReporter(config Config) (sp *StatsdReporter, err error) {

	conn, err := NewUDPConn(config.Address)
	if err != nil {
		return nil, err
	}

	client := NewClient(config.Prefix, conn)

	sp = &StatsdReporter{
		conn:   conn,
		client: client,
	}

	return sp, nil
}

func (sp *StatsdReporter) Close() (err error) {

	err = sp.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sp *StatsdReporter) ReportLogStats(name string, stats log.Stat) (err error) {

	recordCount := stats.EndPosition - stats.StartPosition
	recordCountLabel := fmt.Sprintf(recordCountPattern, name)
	err = sp.client.SetGauge(recordCountLabel, recordCount)
	if err != nil {
		logger.Warn("statsd:", err)
	}

	fileSize := stats.EndOffset - stats.StartOffset
	fileSizeLabel := fmt.Sprintf(fileSizePattern, name)
	err = sp.client.SetGauge(fileSizeLabel, fileSize)
	if err != nil {
		logger.Warn("statsd:", err)
	}

	return nil
}
