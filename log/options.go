package log

import (
	"sync"
)

var (
	DefaultOptions = Options{
		SyncLock: sync.Mutex{},
	}
)

type Options struct {
	SyncLock sync.Mutex
}
