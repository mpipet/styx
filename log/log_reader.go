package log

import (
	"io"
	"sync"

	"gitlab.com/dataptive/styx/recio"
)

type LogReader struct {
	log            *Log
	bufferSize     int
	follow         bool
	ioMode         recio.IOMode
	segmentList    []segmentDescriptor
	currentSegment int
	segmentReader  *segmentReader
	position       int64
	offset         int64
	mustFill       bool
	statChan       chan LogInfo
	logInfo        LogInfo
	closed         bool
	closedLock     sync.RWMutex
}

func newLogReader(l *Log, bufferSize int, follow bool, ioMode recio.IOMode) (lr *LogReader, err error) {

	lr = &LogReader{
		log:            l,
		bufferSize:     bufferSize,
		follow:         follow,
		ioMode:         ioMode,
		segmentList:    []segmentDescriptor{},
		currentSegment: -1,
		segmentReader:  nil,
		position:       0,
		offset:         0,
		mustFill:       false,
		statChan:       make(chan LogInfo, 1),
		logInfo:        LogInfo{},
		closed:         false,
		closedLock:     sync.RWMutex{},
	}

	lr.log.Subscribe(lr.statChan)

	lr.logInfo = <-lr.statChan

	lr.position = lr.logInfo.StartPosition
	lr.offset = lr.logInfo.StartOffset

	return lr, nil
}

func (lr *LogReader) Close() (err error) {

	lr.closedLock.Lock()
	defer lr.closedLock.Unlock()

	if lr.closed {
		return nil
	}

	lr.closed = true

	lr.log.Unsubscribe(lr.statChan)

	close(lr.statChan)

	if lr.segmentReader != nil {
		err = lr.closeCurrentSegment()
		if err != nil {
			return err
		}
	}

	return nil
}

func (lr *LogReader) Tell() (position int64, offset int64) {

	return lr.position, lr.offset
}

func (lr *LogReader) Read(r *Record) (n int, err error) {

	if lr.closed {
		return 0, ErrClosed
	}

Retry:
	if lr.mustFill {
		if lr.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFill
		}

		err = lr.Fill()
		if err != nil {
			return 0, err
		}
	}

	if lr.segmentReader == nil {
		err = lr.openNextSegment()
		if err != nil {
			return 0, err
		}
	}

	n, err = lr.segmentReader.Read(r)

	if err == recio.ErrMustFill {
		lr.mustFill = true

		goto Retry
	}

	if err == io.EOF {
		err = lr.closeCurrentSegment()
		if err != nil {
			return 0, err
		}

		goto Retry
	}

	if err != nil {
		return n, err
	}

	lr.position += 1
	lr.offset += int64(n)

	return n, nil
}

func (lr *LogReader) Fill() (err error) {

Retry:
	if lr.follow && (lr.position >= lr.logInfo.EndPosition) {

		logInfo, more := <-lr.statChan
		if !more {
			return ErrClosed
		}

		lr.logInfo = logInfo

		goto Retry
	}

	lr.closedLock.RLock()
	defer lr.closedLock.RUnlock()

	if lr.closed {
		return ErrClosed
	}

	err = lr.segmentReader.Fill()
	if err != nil {
		return err
	}

	lr.mustFill = false

	return nil
}

func (lr *LogReader) Seek(position int64, whence SeekWhence) (err error) {

	lr.closedLock.RLock()
	defer lr.closedLock.RUnlock()

	if lr.closed {
		return ErrClosed
	}

	var reference int64

	switch whence {
	case SeekOrigin:
		reference = 0

	case SeekStart:
		reference = lr.logInfo.StartPosition

	case SeekCurrent:
		reference = lr.position

	case SeekEnd:
		reference = lr.logInfo.EndPosition

	default:
		return ErrInvalidWhence
	}

	absolute := reference + position

	if absolute < lr.logInfo.StartPosition {
		return ErrOutOfRange
	}

	if absolute > lr.logInfo.EndPosition {
		return ErrOutOfRange
	}

	err = lr.seekPosition(absolute)
	if err != nil {
		return err
	}

	return nil
}

func (lr *LogReader) seekPosition(position int64) (err error) {

	err = lr.updateSegmentList()
	if err != nil {
		return err
	}

	if len(lr.segmentList) == 0 {
		return ErrOutOfRange
	}

	pos := -1
	for i, desc := range lr.segmentList {
		if desc.basePosition > position {
			break
		}
		pos = i
	}

	if pos == -1 {
		return ErrOutOfRange
	}

	current := lr.segmentList[pos]

	segmentReader, err := newSegmentReader(lr.log.path, current.segmentName, lr.bufferSize)

	if err == errSegmentNotExist {
		return ErrOutOfRange
	}

	if err != nil {
		return err
	}

	err = segmentReader.SeekPosition(position)
	if err != nil {
		return err
	}

	position, offset := segmentReader.Tell()

	if lr.segmentReader != nil {
		err = lr.closeCurrentSegment()
		if err != nil {
			return err
		}
	}

	lr.segmentReader = segmentReader
	lr.currentSegment = pos
	lr.position = position
	lr.offset = offset

	return nil
}

func (lr *LogReader) openNextSegment() (err error) {

	nextSegment := lr.currentSegment + 1

	if nextSegment > len(lr.segmentList)-1 {

		err = lr.updateSegmentList()
		if err != nil {
			return err
		}

		if len(lr.segmentList) == 0 {
			return io.EOF
		}

		first := lr.segmentList[0]
		if lr.position < first.basePosition {
			return ErrOutOfRange
		}

		pos := -1
		for i, desc := range lr.segmentList {
			if desc.basePosition == lr.position {
				pos = i
			}
		}

		if pos == -1 {
			return io.EOF
		}

		nextSegment = pos
	}

	next := lr.segmentList[nextSegment]

	segmentReader, err := newSegmentReader(lr.log.path, next.segmentName, lr.bufferSize)

	if err == errSegmentNotExist {
		return ErrOutOfRange
	}

	if err != nil {
		return err
	}

	lr.segmentReader = segmentReader
	lr.currentSegment = nextSegment

	return nil
}

func (lr *LogReader) closeCurrentSegment() (err error) {

	err = lr.segmentReader.Close()
	if err != nil {
		return err
	}

	lr.segmentReader = nil

	return nil
}

func (lr *LogReader) updateSegmentList() (err error) {

	descriptors, err := listSegmentDescriptors(lr.log.path)
	if err != nil {
		return err
	}

	lr.segmentList = descriptors
	lr.currentSegment = -1

	return nil
}
