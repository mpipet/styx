package statsd

import (
	"fmt"
	"net"

	"gitlab.com/dataptive/styx/log"
)

const (
	recordCountPattern = "log.%s.record.count"
	fileSizePattern    = "log.%s.file.size"
)

type StatsdReporter struct {
	conn   net.Conn
	client *Client
}

func NewStatsdReporter(config Config) (sp *StatsdReporter, err error) {

	udpAddr, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
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
	sp.client.Gauge(recordCountLabel, recordCount)

	fileSize := stats.EndOffset - stats.StartOffset
	fileSizeLabel := fmt.Sprintf(fileSizePattern, name)
	sp.client.Gauge(fileSizeLabel, fileSize)

	return nil
}
