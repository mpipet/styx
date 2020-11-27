package log

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/dataptive/styx/clock"
	"gitlab.com/dataptive/styx/recio"
)

type SyncMode string

const (
	SyncManual SyncMode = "manual" // Sync is called manually by the user.
	SyncAuto   SyncMode = "auto"   // A Sync is automatically queued after each Flush.
	SyncUnsafe SyncMode = "unsafe" // Readers are notified after Flush without needing a Sync.
)

type SeekWhence string

const (
	SeekOrigin  SeekWhence = "origin"  // Seek from the log origin (position 0).
	SeekStart   SeekWhence = "start"   // Seek from the first available record.
	SeekCurrent SeekWhence = "current" // Seek from the current position.
	SeekEnd     SeekWhence = "end"     // Seek from the end of the log.
)

var (
	ErrExist          = errors.New("log: log already exists")
	ErrNotExist       = errors.New("log: log does not exist")
	ErrUnknownVersion = errors.New("log: unknown version")
	ErrConfigCorrupt  = errors.New("log: config corrupt")
	ErrOutOfRange     = errors.New("log: position out of range")
	ErrInvalidWhence  = errors.New("log: invalid whence")
)

const (
	configFilename = "config"

	dirPerm  = 0744
	filePerm = 0644
)

var (
	now = clock.New(time.Second)
)

type LogInfo struct {
	StartPosition  int64
	StartOffset    int64
	StartTimestamp int64
	EndPosition    int64
	EndOffset      int64
}

type Log struct {
	path        string
	config      Config
	options     Options
	logInfo     LogInfo
	subscribers []chan LogInfo
	notifyLock  sync.Mutex
}

func Create(path string, config Config, options Options) (l *Log, err error) {

	err = os.Mkdir(path, os.FileMode(dirPerm))
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrExist
		}

		return nil, err
	}

	// Sync parent directory
	parentPath := filepath.Dir(path)
	err = syncDirectory(parentPath)
	if err != nil {
		return nil, err
	}

	// Store config file
	pathname := filepath.Join(path, configFilename)
	err = config.dump(pathname)
	if err != nil {
		return nil, err
	}

	// Sync config file
	err = syncFile(pathname)
	if err != nil {
		return nil, err
	}

	// Sync log directory
	err = syncDirectory(path)
	if err != nil {
		return nil, err
	}

	l, err = newLog(path, config, options)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func Open(path string, options Options) (l *Log, err error) {

	config := Config{}

	// Load config file
	pathname := filepath.Join(path, configFilename)

	err = config.load(pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotExist
		}

		return nil, err
	}

	l, err = newLog(path, config, options)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func Delete(path string) (err error) {

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	parentPath := filepath.Dir(path)

	err = syncDirectory(parentPath)
	if err != nil {
		return err
	}

	return nil
}

func newLog(path string, config Config, options Options) (l *Log, err error) {

	l = &Log{
		path:        path,
		config:      config,
		options:     options,
		logInfo:     LogInfo{},
		subscribers: []chan LogInfo{},
		notifyLock:  sync.Mutex{},
	}

	lw, err := l.NewWriter(0, SyncManual, recio.ModeAuto)
	if err != nil {
		return nil, err
	}

	err = lw.Close()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Log) NewWriter(bufferSize int, syncMode SyncMode, ioMode recio.IOMode) (lw *LogWriter, err error) {

	lw, err = newLogWriter(l, bufferSize, syncMode, ioMode)
	if err != nil {
		return nil, err
	}

	return lw, nil
}

func (l *Log) NewReader(bufferSize int, follow bool, ioMode recio.IOMode) (lr *LogReader, err error) {

	lr, err = newLogReader(l, bufferSize, follow, ioMode)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (l *Log) notify(logInfo LogInfo) {

	l.notifyLock.Lock()
	defer l.notifyLock.Unlock()

	l.logInfo = logInfo

	for _, subscriber := range l.subscribers {
		select {
		case <-subscriber:
		default:
		}
		subscriber <- logInfo
	}
}

func (l *Log) Stat() (logInfo LogInfo) {

	l.notifyLock.Lock()
	defer l.notifyLock.Unlock()

	logInfo = l.logInfo

	return logInfo
}

func (l *Log) Subscribe(subscriber chan LogInfo) {

	l.notifyLock.Lock()
	defer l.notifyLock.Unlock()

	select {
	case <-subscriber:
	default:
	}

	subscriber <- l.logInfo

	l.subscribers = append(l.subscribers, subscriber)
}

func (l *Log) Unsubscribe(subscriber chan LogInfo) {

	l.notifyLock.Lock()
	defer l.notifyLock.Unlock()

	pos := -1
	for i, s := range l.subscribers {
		if s == subscriber {
			pos = i
			break
		}
	}

	if pos == -1 {
		return
	}

	l.subscribers[pos] = l.subscribers[len(l.subscribers)-1]
	l.subscribers = l.subscribers[:len(l.subscribers)-1]
}
