package manager

import (
	"errors"
	"io"
	"path/filepath"
	"sync"

	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/log"
)

var (
	ErrNotExist    = errors.New("manager: log does not exist")
	ErrUnavailable = errors.New("manager: log unavailable")
)

type LogManager struct {
	config   Config
	logs     []*Log
	logsLock sync.Mutex
}

func NewLogManager(config Config) (lm *LogManager, err error) {

	logger.Debugf("manager: starting log manager (data_directory=%s)", config.DataDirectory)

	lm = &LogManager{
		config: config,
	}

	names, err := listLogs(lm.config.DataDirectory)
	if err != nil {
		return nil, err
	}

	for _, name := range names {

		logger.Debugf("manager: opening log %s", name)

		ml, err := openLog(lm.config.DataDirectory, name, log.DefaultOptions, lm.config.WriteBufferSize)
		if err != nil {
			return lm, err
		}

		lm.logs = append(lm.logs, ml)

		if ml.Status() != StatusOK {

			logger.Debugf("manager: scanning log %s", name)

			go ml.scan()
		}
	}

	return lm, nil
}

func (lm *LogManager) Close() (err error) {

	logger.Debugf("manager: closing log manager")

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	for _, ml := range lm.logs {

		err = ml.close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (lm *LogManager) ListLogs() (logs []*Log) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	logs = lm.logs

	return logs
}

func (lm *LogManager) CreateLog(name string, logConfig log.Config) (ml *Log, err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	ml, err = createLog(lm.config.DataDirectory, name, logConfig, log.DefaultOptions, lm.config.WriteBufferSize)
	if err != nil {
		return nil, err
	}

	lm.logs = append(lm.logs, ml)

	return ml, nil
}

func (lm *LogManager) GetLog(name string) (ml *Log, err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	found := false
	for _, current := range lm.logs {

		if current.name == name {
			ml = current
			found = true
			break
		}
	}

	if !found {
		return nil, ErrNotExist
	}

	return ml, nil
}

func (lm *LogManager) DeleteLog(name string) (err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	pos := -1
	for i, ml := range lm.logs {
		if ml.name == name {
			pos = i
			break
		}
	}

	if pos == -1 {
		return ErrNotExist
	}

	ml := lm.logs[pos]

	err = ml.close()
	if err != nil {
		return err
	}

	lm.logs[pos] = lm.logs[len(lm.logs)-1]
	lm.logs = lm.logs[:len(lm.logs)-1]

	path := filepath.Join(lm.config.DataDirectory, name)

	err = log.Delete(path)
	if err != nil {
		return err
	}

	return nil
}

func (lm *LogManager) RestoreLog(name string, r io.Reader) (err error) {

	pathname := filepath.Join(lm.config.DataDirectory, name)

	err = log.Restore(pathname, r)
	if err != nil {
		return err
	}

	ml, err := openLog(lm.config.DataDirectory, name, log.DefaultOptions, lm.config.WriteBufferSize)
	if err != nil {
		return err
	}

	lm.logsLock.Lock()
	lm.logs = append(lm.logs, ml)
	lm.logsLock.Unlock()

	return nil
}

func listLogs(path string) (names []string, err error) {

	pattern := path + "/*"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		_, filename := filepath.Split(match)
		names = append(names, filename)
	}

	return names, nil
}