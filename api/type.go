package api

import (
	"errors"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logman"
)

const (
	TimeoutHeaderName     = "X-Styx-Timeout"
	RecordLinesMediaType  = "application/ld+text"
	RecordBinaryMediaType = "application/vnd.styx-records"
	RaftProtocolString    = "hashicorp-raft/3"
	StyxProtocolString    = "styx/0"
)

var (
	ErrInvalidWhence = errors.New("invalid whence")
)

type LogInfo struct {
	Name          string           `json:"name"`
	Status        logman.LogStatus `json:"status"`
	RecordCount   int64            `json:"record_count"`
	FileSize      int64            `json:"file_size"`
	StartPosition int64            `json:"start_position"`
	EndPosition   int64            `json:"end_position"`
}

type LogConfig struct {
	MaxRecordSize   int   `schema:"max_record_size"`
	IndexAfterSize  int64 `schema:"index_after_size"`
	SegmentMaxCount int64 `schema:"segment_max_count"`
	SegmentMaxSize  int64 `schema:"segment_max_size"`
	SegmentMaxAge   int64 `schema:"segment_max_age"`
	LogMaxCount     int64 `schema:"log_max_count"`
	LogMaxSize      int64 `schema:"log_max_size"`
	LogMaxAge       int64 `schema:"log_max_age"`
}

type ListLogsResponse []LogInfo

type CreateLogForm struct {
	Name string `schema:"name,required"`
	*LogConfig
}
type CreateLogResponse LogInfo

type GetLogResponse LogInfo

type RestoreLogParams struct {
	Name string `schema:"name,required"`
}

type WriteRecordResponse struct {
	Position int64 `json:"position"`
	Count    int64 `json:"count"`
}

type ReadRecordParams struct {
	Whence   log.Whence `schema:"whence"`
	Position int64      `schema:"position"`
}

func (p ReadRecordParams) Validate() (err error) {
	err = validateWhence(p.Whence)

	if err != nil {
		return err
	}

	return nil
}

type WriteRecordsBatchResponse WriteRecordResponse

type ReadRecordsBatchParams struct {
	Whence   log.Whence `schema:"whence"`
	Position int64      `schema:"position"`
	Count    int64      `schema:"count"`
	Longpoll bool       `schema:"longpoll"`
}

func (p ReadRecordsBatchParams) Validate() (err error) {
	err = validateWhence(p.Whence)

	if err != nil {
		return err
	}

	return nil
}

type WriteRecordsLinesResponse WriteRecordResponse

type ReadRecordsLinesParams ReadRecordsBatchParams

func (p ReadRecordsLinesParams) Validate() (err error) {
	err = validateWhence(p.Whence)

	if err != nil {
		return err
	}

	return nil
}

type ReadRecordsTCPParams struct {
	Whence   log.Whence `schema:"whence"`
	Position int64      `schema:"position"`
	Count    int64      `schema:"count"`
	Follow   bool       `schema:"follow"`
}

func (p ReadRecordsTCPParams) Validate() (err error) {
	err = validateWhence(p.Whence)

	if err != nil {
		return err
	}

	return nil
}

type ReadRecordsWSParams ReadRecordsTCPParams

func (p ReadRecordsWSParams) Validate() (err error) {
	err = validateWhence(p.Whence)

	if err != nil {
		return err
	}

	return nil
}

func validateWhence(whence log.Whence) (err error) {

	validWhences := []log.Whence{
		log.SeekOrigin,
		log.SeekStart,
		log.SeekCurrent,
		log.SeekEnd,
	}

	found := false
	for _, w := range validWhences {

		if w == whence {
			found = true
			break
		}
	}

	if !found {
		return ErrInvalidWhence
	}

	return nil
}

type ListNodesResponse []Node

type Node struct {
	Name     string `json:"name"`
	Leader   bool   `json:"leader"`
	Suffrage string `json:"suffrage"`
	Address  string `json:"address"`
}

type AddNodeForm struct {
	Name    string `schema:"name,required"`
	Address string `schema:"address,required"`
	Voter   bool   `schema:"voter"`
}
