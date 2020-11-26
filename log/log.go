package log

import (
	"errors"
	"time"

	"gitlab.com/dataptive/styx/clock"
)

var (
	ErrOutOfRange = errors.New("log: position out of range")

	now = clock.New(time.Second)
)

type Config struct {
	IndexAfterSize  int64
	MaxRecordSize   int
	SegmentMaxCount int64
	SegmentMaxSize  int64
	SegmentMaxAge   int64
}
