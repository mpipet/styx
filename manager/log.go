package manager

import (
	"io"
	"path/filepath"
	"sync"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
)

type LogStatus string

const (
	StatusOK       LogStatus = "ok"
	StatusCorrupt  LogStatus = "corrupt"
	StatusTainted  LogStatus = "tainted"
	StatusScanning LogStatus = "scanning"
	StatusUnknown  LogStatus = "unknown"
)

type LogInfo struct {
	Name          string
	Status        LogStatus
	RecordCount   int64
	FileSize      int64
	StartPosition int64
	EndPosition   int64
}

type Log struct {
	path             string
	name             string
	options          log.Options
	writerBufferSize int
	status           LogStatus
	log              *log.Log
	writer           *log.LogWriter
	fanin            *log.Fanin
	lock             sync.RWMutex
}

func (ml *Log) NewWriter(ioMode recio.IOMode) (fw *log.FaninWriter, err error) {

	if ml.Status() != StatusOK {
		return nil, ErrUnavailable
	}

	fw = log.NewFaninWriter(ml.fanin, ioMode)

	return fw, nil
}

func (ml *Log) NewReader(bufferSize int, follow bool, ioMode recio.IOMode) (lr *log.LogReader, err error) {

	if ml.Status() != StatusOK {
		return nil, ErrUnavailable
	}

	lr, err = ml.log.NewReader(bufferSize, follow, ioMode)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (ml *Log) Status() (status LogStatus) {

	ml.lock.RLock()
	defer ml.lock.RUnlock()
	status = ml.status

	return status
}

func (ml *Log) Stat() (logInfo LogInfo) {

	status := ml.Status()

	if status != StatusOK {
		logInfo = LogInfo{
			Name:   ml.name,
			Status: ml.status,
		}

		return logInfo
	}

	fileInfo := ml.log.Stat()

	recordCount := fileInfo.EndPosition - fileInfo.StartPosition
	fileSize := fileInfo.EndOffset - fileInfo.StartOffset

	logInfo = LogInfo{
		Name:          ml.name,
		Status:        status,
		RecordCount:   recordCount,
		FileSize:      fileSize,
		StartPosition: fileInfo.StartPosition,
		EndPosition:   fileInfo.EndPosition,
	}

	return logInfo
}

func (ml *Log) Backup(w io.Writer) (err error) {

	if ml.Status() != StatusOK {
		return ErrUnavailable
	}

	err = ml.log.Backup(w)
	if err != nil {
		return err
	}

	return nil
}

func createLog(path, name string, config log.Config, options log.Options, writerBufferSize int) (ml *Log, err error) {

	ml = &Log{
		path:             path,
		name:             name,
		options:          options,
		writerBufferSize: writerBufferSize,
		status:           StatusUnknown,
	}

	pathname := filepath.Join(path, name)

	l, err := log.Create(pathname, config, options)
	if err != nil {
		return nil, err
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		return nil, err
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)

	return ml, nil
}

func openLog(path, name string, options log.Options, writerBufferSize int) (ml *Log, err error) {

	ml = &Log{
		path:             path,
		name:             name,
		options:          options,
		writerBufferSize: writerBufferSize,
		status:           StatusUnknown,
	}

	pathname := filepath.Join(path, name)

	l, err := log.Open(pathname, options)
	if err != nil {

		// TODO return err not exists (or other kin of errors)??

		ml.status = StatusTainted

		if err == log.ErrCorrupt {
			ml.status = StatusCorrupt
		}

		return ml, nil
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		return nil, err
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)

	return ml, nil
}

func (ml *Log) close() (err error) {

	if ml.Status() != StatusOK {
		return nil
	}

	err = ml.fanin.Close()
	if err != nil {
		return err
	}

	err = ml.writer.Close()
	if err != nil {
		return err
	}

	err = ml.log.Close()
	if err != nil {
		return err
	}

	return nil
}

func (ml *Log) scan() {

	pathname := filepath.Join(ml.path, ml.name)

	ml.lock.Lock()
	ml.status = StatusScanning

	// Make log unavailable during scan.
	if ml.log != nil {
		err := ml.log.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	if ml.writer != nil {
		err := ml.writer.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	if ml.fanin != nil {
		err := ml.fanin.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	ml.lock.Unlock()

	// Perform log scan.
	err := log.Scan(pathname)

	ml.lock.Lock()
	defer ml.lock.Unlock()

	if err != nil {

		ml.status = StatusTainted

		if err == log.ErrCorrupt {
			ml.status = StatusCorrupt
		}

		return
	}

	// Try to make log functionnal again.
	l, err := log.Open(pathname, ml.options)

	if err == log.ErrCorrupt {
		ml.status = StatusCorrupt
		return
	}

	if err != nil {
		ml.status = StatusTainted
		return
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		ml.status = StatusTainted
		return
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)
}

