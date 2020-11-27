package log

import (
	"sync"
	"time"

	"gitlab.com/dataptive/styx/recio"
)

const (
	expireInterval  = time.Second
	maxDirtyWriters = 50
)

type LogWriter struct {
	log             *Log
	bufferSize      int
	syncMode        SyncMode
	ioMode          recio.IOMode
	segmentList     []segmentDescriptor
	segmentLock     sync.Mutex
	segmentWriter   *segmentWriter
	buffered        int64
	position        int64
	offset          int64
	mustFlush       bool
	mustRoll        bool
	expirerStop     chan struct{}
	expirerDone     chan struct{}
	currentName     string
	currentDirty    bool
	directoryDirty  bool
	dirtyList       []string
	flushedPosition int64
	flushedOffset   int64
	dirtyLock       sync.Mutex
	syncerChan      chan struct{}
	syncerDone      chan struct{}
	logInfo         LogInfo
}

func newLogWriter(l *Log, bufferSize int, syncMode SyncMode, ioMode recio.IOMode) (lw *LogWriter, err error) {

	lw = &LogWriter{
		log:             l,
		bufferSize:      bufferSize,
		syncMode:        syncMode,
		ioMode:          ioMode,
		segmentList:     []segmentDescriptor{},
		segmentLock:     sync.Mutex{},
		buffered:        0,
		position:        0,
		offset:          0,
		mustFlush:       false,
		mustRoll:        false,
		expirerStop:     make(chan struct{}),
		expirerDone:     make(chan struct{}),
		currentName:     "",
		currentDirty:    false,
		directoryDirty:  false,
		dirtyList:       []string{},
		flushedPosition: 0,
		flushedOffset:   0,
		dirtyLock:       sync.Mutex{},
		syncerChan:      make(chan struct{}, 1),
		syncerDone:      make(chan struct{}),
		logInfo:         LogInfo{},
	}

	err = lw.updateSegmentList()
	if err != nil {
		return nil, err
	}

	if len(lw.segmentList) > 0 {
		err = lw.openLastSegment()
		if err != nil {
			return nil, err
		}
	} else {
		err = lw.createNewSegment()
		if err != nil {
			return nil, err
		}
	}

	lw.flushedPosition = lw.position
	lw.flushedOffset = lw.offset

	first := lw.segmentList[0]
	lw.logInfo = LogInfo{
		StartPosition:  first.basePosition,
		StartOffset:    first.baseOffset,
		StartTimestamp: first.baseTimestamp,
		EndPosition:    lw.flushedPosition,
		EndOffset:      lw.flushedOffset,
	}
	lw.log.notify(lw.logInfo)

	go lw.expirer()
	go lw.syncer()

	return lw, nil
}

func (lw *LogWriter) expirer() {

	ticker := time.NewTicker(expireInterval)

	for {
		select {
		case <-lw.expirerStop:
			ticker.Stop()
			lw.expirerDone <- struct{}{}
			return
		case <-ticker.C:
			err := lw.enforceMaxAge()
			if err != nil {
				panic(err)
			}
		}
	}
}

func (lw *LogWriter) syncer() {

	for _ = range lw.syncerChan {
		err := lw.Sync()
		if err != nil {
			panic(err)
		}
	}

	lw.syncerDone <- struct{}{}
}

func (lw *LogWriter) Close() (err error) {

	lw.expirerStop <- struct{}{}
	<-lw.expirerDone

	close(lw.syncerChan)
	<-lw.syncerDone

	err = lw.closeCurrentSegment()
	if err != nil {
		return err
	}

	return nil
}

func (lw *LogWriter) Tell() (position int64, offset int64) {

	return lw.position, lw.offset
}

func (lw *LogWriter) Write(r *Record) (n int, err error) {

Retry:
	if lw.mustFlush {
		if lw.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFlush
		}

		err = lw.Flush()
		if err != nil {
			return 0, err
		}
	}

	if lw.mustRoll {
		err = lw.closeCurrentSegment()
		if err != nil {
			return 0, err
		}

		err = lw.createNewSegment()
		if err != nil {
			return 0, err
		}

		lw.mustRoll = false
	}

	n, err = lw.segmentWriter.Write(r)

	if err == recio.ErrMustFlush {
		lw.mustFlush = true

		goto Retry
	}

	if err == errSegmentFull {
		lw.mustFlush = true
		lw.mustRoll = true

		goto Retry
	}

	if err != nil {
		return n, err
	}

	lw.buffered += 1
	lw.position += 1
	lw.offset += int64(n)

	return n, nil
}

func (lw *LogWriter) Flush() (err error) {

	err = lw.enforceMaxCount()
	if err != nil {
		return err
	}

	err = lw.enforceMaxSize()
	if err != nil {
		return err
	}

	err = lw.enforceMaxAge()
	if err != nil {
		return err
	}

	err = lw.segmentWriter.Flush()
	if err != nil {
		return err
	}

	lw.buffered = 0
	lw.mustFlush = false

	lw.dirtyLock.Lock()
	defer lw.dirtyLock.Unlock()

	lw.currentDirty = true
	lw.flushedPosition = lw.position
	lw.flushedOffset = lw.offset

	forceSync := false
	if len(lw.dirtyList) >= maxDirtyWriters {
		forceSync = true
	}

	if lw.syncMode == SyncAuto {
		if !forceSync {
			select {
			case <-lw.syncerChan:
			default:
			}
		}
		lw.syncerChan <- struct{}{}
	}

	if lw.syncMode == SyncManual {
		if forceSync {
			lw.syncerChan <- struct{}{}
		}
	}

	if lw.syncMode == SyncUnsafe {
		lw.logInfo.EndPosition = lw.flushedPosition
		lw.logInfo.EndOffset = lw.flushedOffset
		lw.log.notify(lw.logInfo)
	}

	return nil
}

func (lw *LogWriter) Sync() (err error) {

	if lw.syncMode == SyncUnsafe {
		return nil
	}

	lw.log.options.SyncLock.Lock()
	defer lw.log.options.SyncLock.Unlock()

	lw.dirtyLock.Lock()

	currentName := lw.currentName
	currentDirty := lw.currentDirty
	directoryDirty := lw.directoryDirty
	dirtyList := lw.dirtyList
	flushedPosition := lw.flushedPosition
	flushedOffset := lw.flushedOffset

	lw.currentDirty = false
	lw.directoryDirty = false
	lw.dirtyList = lw.dirtyList[:0]

	lw.dirtyLock.Unlock()

	if len(dirtyList) > 0 {
		for _, segmentName := range dirtyList {
			err = syncSegment(lw.log.path, segmentName)
			if err != nil {
				return err
			}
		}
	}

	if directoryDirty {
		err = syncDirectory(lw.log.path)
		if err != nil {
			return err
		}
	}

	if currentDirty {
		err = syncSegment(lw.log.path, currentName)
		if err != nil {
			return err
		}
	}

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	lw.logInfo.EndPosition = flushedPosition
	lw.logInfo.EndOffset = flushedOffset
	lw.log.notify(lw.logInfo)

	return nil
}

func (lw *LogWriter) enforceMaxCount() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	if lw.log.config.LogMaxCount == -1 {
		return nil
	}

	if len(lw.segmentList) <= 1 {
		return nil
	}

	expiredPosition := lw.position - lw.log.config.LogMaxCount

	// Last segment should never be deleted.
	descriptors := lw.segmentList[:len(lw.segmentList)-1]

	pos := 0
	for _, desc := range descriptors {
		if desc.basePosition >= expiredPosition {
			break
		}
		pos += 1
	}

	if pos == 0 {
		return nil
	}

	deleteList := lw.segmentList[:pos]
	lw.segmentList = lw.segmentList[pos:]

	err = lw.deleteSegments(deleteList)
	if err != nil {
		return err
	}

	first := lw.segmentList[0]
	lw.logInfo.StartPosition = first.basePosition
	lw.logInfo.StartOffset = first.baseOffset
	lw.logInfo.StartTimestamp = first.baseTimestamp
	lw.log.notify(lw.logInfo)

	return nil
}

func (lw *LogWriter) enforceMaxSize() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	if lw.log.config.LogMaxSize == -1 {
		return nil
	}

	if len(lw.segmentList) <= 1 {
		return nil
	}

	expiredOffset := lw.offset - lw.log.config.LogMaxSize

	// Last segment should never be deleted.
	descriptors := lw.segmentList[:len(lw.segmentList)-1]

	pos := 0
	for _, desc := range descriptors {
		if desc.baseOffset >= expiredOffset {
			break
		}
		pos += 1
	}

	if pos == 0 {
		return nil
	}

	deleteList := lw.segmentList[:pos]
	lw.segmentList = lw.segmentList[pos:]

	err = lw.deleteSegments(deleteList)
	if err != nil {
		return err
	}

	first := lw.segmentList[0]
	lw.logInfo.StartPosition = first.basePosition
	lw.logInfo.StartOffset = first.baseOffset
	lw.logInfo.StartTimestamp = first.baseTimestamp
	lw.log.notify(lw.logInfo)

	return nil
}

func (lw *LogWriter) enforceMaxAge() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	if lw.log.config.LogMaxAge == -1 {
		return nil
	}

	if len(lw.segmentList) <= 1 {
		return nil
	}

	timestamp := now.Unix()

	expiredTimestamp := timestamp - lw.log.config.LogMaxAge

	// Last segment should never be deleted.
	descriptors := lw.segmentList[:len(lw.segmentList)-1]

	pos := 0
	for _, desc := range descriptors {
		if desc.baseTimestamp >= expiredTimestamp {
			break
		}
		pos += 1
	}

	if pos == 0 {
		return nil
	}

	deleteList := lw.segmentList[:pos]
	lw.segmentList = lw.segmentList[pos:]

	err = lw.deleteSegments(deleteList)
	if err != nil {
		return err
	}

	first := lw.segmentList[0]
	lw.logInfo.StartPosition = first.basePosition
	lw.logInfo.StartOffset = first.baseOffset
	lw.logInfo.StartTimestamp = first.baseTimestamp
	lw.log.notify(lw.logInfo)

	return nil
}

func (lw *LogWriter) deleteSegments(descriptors []segmentDescriptor) (err error) {

	if len(descriptors) == 0 {
		return nil
	}

	for _, desc := range descriptors {
		err = deleteSegment(lw.log.path, desc.segmentName)

		if err == errSegmentNotExist {
			continue
		}

		if err != nil {
			return err
		}
	}

	if lw.syncMode == SyncUnsafe {
		return nil
	}

	lw.dirtyLock.Lock()
	defer lw.dirtyLock.Unlock()

	lw.directoryDirty = true

	return nil
}

func (lw *LogWriter) createNewSegment() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	timestamp := now.Unix()

	name := buildSegmentName(lw.position, lw.offset, timestamp)
	desc := segmentDescriptor{
		segmentName:   name,
		basePosition:  lw.position,
		baseOffset:    lw.offset,
		baseTimestamp: timestamp,
	}

	segmentWriter, err := newSegmentWriter(lw.log.path, desc.segmentName, true, lw.log.config, lw.bufferSize)
	if err != nil {
		return err
	}

	lw.segmentWriter = segmentWriter
	lw.segmentList = append(lw.segmentList, desc)

	if lw.syncMode == SyncUnsafe {
		return nil
	}

	lw.dirtyLock.Lock()
	defer lw.dirtyLock.Unlock()

	lw.currentName = desc.segmentName
	lw.currentDirty = false
	lw.directoryDirty = true

	return nil
}

func (lw *LogWriter) openLastSegment() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	last := lw.segmentList[len(lw.segmentList)-1]

	segmentWriter, err := newSegmentWriter(lw.log.path, last.segmentName, false, lw.log.config, lw.bufferSize)
	if err != nil {
		return err
	}

	position, offset := segmentWriter.Tell()

	lw.segmentWriter = segmentWriter
	lw.position = position
	lw.offset = offset

	if lw.syncMode == SyncUnsafe {
		return nil
	}

	lw.dirtyLock.Lock()
	defer lw.dirtyLock.Unlock()

	lw.currentName = last.segmentName
	lw.currentDirty = false
	lw.directoryDirty = false

	return nil
}

func (lw *LogWriter) closeCurrentSegment() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	err = lw.segmentWriter.Close()
	if err != nil {
		return err
	}

	lw.segmentWriter = nil

	// Delete segment if empty.
	current := lw.segmentList[len(lw.segmentList)-1]

	if current.basePosition == (lw.position - lw.buffered) {
		err = deleteSegment(lw.log.path, current.segmentName)
		if err != nil {
			return err
		}
	}

	lw.buffered = 0

	if lw.syncMode == SyncUnsafe {
		return nil
	}

	lw.dirtyLock.Lock()
	defer lw.dirtyLock.Unlock()

	if lw.currentDirty {
		lw.dirtyList = append(lw.dirtyList, lw.currentName)
	}

	lw.currentName = ""
	lw.currentDirty = false

	return nil
}

func (lw *LogWriter) updateSegmentList() (err error) {

	lw.segmentLock.Lock()
	defer lw.segmentLock.Unlock()

	descriptors, err := listSegmentDescriptors(lw.log.path)
	if err != nil {
		return err
	}

	lw.segmentList = descriptors

	return nil
}
