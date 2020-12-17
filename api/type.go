package api

import (
	"gitlab.com/dataptive/styx/manager"
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
