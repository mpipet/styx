package log

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/dataptive/styx/clock"
	"gitlab.com/dataptive/styx/flock"
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

type SyncHandler func(position int64)

type ErrorHandler func(err error)

var (
	ErrExist          = errors.New("log: already exists")
	ErrNotExist       = errors.New("log: does not exist")
	ErrUnknownVersion = errors.New("log: unknown version")
	ErrConfigCorrupt  = errors.New("log: config corrupt")
	ErrOutOfRange     = errors.New("log: position out of range")
	ErrInvalidWhence  = errors.New("log: invalid whence")
	ErrLocked         = errors.New("log: already locked")
	ErrOrphaned       = errors.New("log: orphaned")
)

const (
	configFilename = "config"
	lockFilename   = "lock"

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
	writerLock  sync.Mutex
	lockFile    *flock.Flock
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

func Restore(path string, r io.Reader) (err error) {

	err = os.Mkdir(path, os.FileMode(dirPerm))
	if err != nil {
		if os.IsExist(err) {
			return ErrExist
		}

		return err
	}

	parentDirname := filepath.Dir(path)

	err = syncDirectory(parentDirname)
	if err != nil {
		panic(err)
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		pathname := filepath.Join(path, header.Name)

		f, err := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE, os.FileMode(header.Mode))
		if err != nil {
			return err
		}

		err = syncDirectory(path)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(f, tr)
		if err != nil {
			return err
		}

		err = f.Sync()
		if err != nil {
			panic(err)
		}
	}

	err = gzr.Close()
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
		writerLock:  sync.Mutex{},
		lockFile:    nil,
	}

	err = l.acquireFileLock()
	if err != nil {
		return nil, err
	}

	err = l.initializeStat()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Log) Close() (err error) {

	err = l.releaseFileLock()
	if err != nil {
		return err
	}

	return nil
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

func (l *Log) Backup(w io.Writer) (err error) {

	// Checkpoint current log state.
	logInfo := l.Stat()

	// Build a list a index and records file handles.
	names, err := listSegments(l.path)
	if err != nil {
		return err
	}

	var recordsFiles []*os.File
	var indexFiles []*os.File

	for _, name := range names {

		pathname := filepath.Join(l.path, name)

		f, err := os.Open(pathname + recordsSuffix)
		if err != nil {
			return err
		}

		recordsFiles = append(recordsFiles, f)

		f, err = os.Open(pathname + indexSuffix)
		if err != nil {
			return err
		}

		indexFiles = append(indexFiles, f)
	}

	// Get a config file handle.
	configPathname := filepath.Join(l.path, configFilename)

	configFile, err := os.Open(configPathname)
	if err != nil {
		return err
	}

	// Prepare the tar gz writer.
	gzw := gzip.NewWriter(w)
	tw := tar.NewWriter(gzw)

	// Add the config file to the archive.
	fi, err := configFile.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, configFile)
	if err != nil {
		return err
	}

	err = configFile.Close()
	if err != nil {
		return err
	}

	// Save the last records file to process it separately.
	lastRecordsFile := recordsFiles[len(recordsFiles)-1]

	// Add all records files except the last one to the archive.
	recordsFiles = recordsFiles[:len(recordsFiles)-1]

	for _, recordsFile := range recordsFiles {

		fi, err := recordsFile.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, recordsFile)
		if err != nil {
			return err
		}

		err = recordsFile.Close()
		if err != nil {
			return err
		}
	}

	// Copy the last records file to the archive, up to the checkpointed
	// offset.
	fi, err = lastRecordsFile.Stat()
	if err != nil {
		return err
	}

	filename := fi.Name()
	segmentName := filename[:len(filename)-len(recordsSuffix)]
	_, baseOffset, _ := parseSegmentName(segmentName)

	header = &tar.Header{
		Name: fi.Name(),
		Mode: int64(fi.Mode().Perm()),
		Size: logInfo.EndOffset - baseOffset,
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	lr := &io.LimitedReader{
		R: lastRecordsFile,
		N: header.Size,
	}

	_, err = io.Copy(tw, lr)
	if err != nil {
		return err
	}

	err = lastRecordsFile.Close()
	if err != nil {
		return err
	}

	// Save the last index file to process it separately.
	lastIndexFile := indexFiles[len(indexFiles)-1]

	// Add all index files except the last one to the archive.
	indexFiles = indexFiles[:len(indexFiles)-1]

	for _, indexFile := range indexFiles {

		fi, err := indexFile.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, indexFile)
		if err != nil {
			return err
		}

		err = indexFile.Close()
		if err != nil {
			return err
		}
	}

	// Find the last index entry matching the checkpointed state and copy
	// the last index file up to this point.
	indexBufferedReader := recio.NewBufferedReader(lastIndexFile, 1 << 20, recio.ModeAuto)
	indexAtomicReader := recio.NewAtomicReader(indexBufferedReader)

	ie := indexEntry{}

	offset := int64(0)
	for {
		n, err := indexAtomicReader.Read(&ie)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if ie.position > logInfo.EndOffset {
			break
		}

		offset += int64(n)
	}

	fi, err = lastIndexFile.Stat()
	if err != nil {
		return err
	}

	header = &tar.Header{
		Name: fi.Name(),
		Mode: int64(fi.Mode().Perm()),
		Size: offset,
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = lastIndexFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	lr = &io.LimitedReader{
		R: lastIndexFile,
		N: header.Size,
	}

	_, err = io.Copy(tw, lr)
	if err != nil {
		return err
	}

	err = lastIndexFile.Close()
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	err = gzw.Close()
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) initializeStat() (err error) {

	lw, err := l.NewWriter(0, SyncManual, recio.ModeAuto)
	if err != nil {
		return err
	}

	err = lw.Close()
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) acquireWriterLock() {

	l.writerLock.Lock()
}

func (l *Log) releaseWriterLock() {

	l.writerLock.Unlock()
}

func (l *Log) acquireFileLock() (err error) {

	pathname := filepath.Join(l.path, lockFilename)

	l.lockFile = flock.New(pathname, os.FileMode(filePerm))

	err = l.lockFile.Acquire()

	if err == flock.ErrLocked {
		return ErrLocked
	}

	if err == flock.ErrOrphaned {
		l.lockFile.Clear()
		return ErrOrphaned
	}

	if err != nil {
		return err
	}

	return nil
}

func (l *Log) releaseFileLock() (err error) {

	err = l.lockFile.Release()
	if err != nil {
		return err
	}

	return nil
}
