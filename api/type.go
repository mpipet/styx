package api

import (
	"errors"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/manager"
)

var (
	ErrInvalidWhence = errors.New("invalid whence")
)

type LogInfo struct {
	Name          string            `json:"name"`
	Status        manager.LogStatus `json:"status"`
	RecordCount   int64             `json:"record_count"`
	FileSize      int64             `json:"file_size"`
	StartPosition int64             `json:"start_position"`
	EndPosition   int64             `json:"end_position"`
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
