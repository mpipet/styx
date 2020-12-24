package nodeman

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sync"

	"gitlab.com/dataptive/styx/logger"

	"github.com/hashicorp/raft"
)

var (
	ErrNotFound = errors.New("nodeman: key not found")
)

type FSMCommand struct {
	Operation string
	Key string
	Value interface{}
}

type FSMResult struct {
	Value interface{}
	Err error
}

type FSM struct {
	state map[string]interface{}
	stateLock sync.Mutex
}

func newFSM() (fsm *FSM) {

	fsm = &FSM{
		state:     make(map[string]interface{}),
		stateLock: sync.Mutex{},
	}

	return fsm
}

func (f *FSM) GetState() (state map[string]interface{}) {

	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	state = f.state

	return state
}

func (f *FSM) Apply(log *raft.Log) (applier interface{}) {

	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	var command FSMCommand

	err := json.Unmarshal(log.Data, &command)
	if err != nil {
		panic(err)
	}

	var result FSMResult

	switch command.Operation {
	case "get":
		value, ok := f.state[command.Key]
		if !ok {
			result.Err = ErrNotFound
			return &result
		}

		result.Value = value
		return &result
	case "set":
		f.state[command.Key] = command.Value
		logger.Tracef("state updated: %+v", f.state)
		return &result
	case "delete":
		delete(f.state, command.Key)
		logger.Tracef("state updated: %+v", f.state)
		return &result
	case "list":
		result.Value = f.state
		return &result
	default:
		//
	}

	return nil
}

func (f *FSM) Snapshot() (fsmSnapshot raft.FSMSnapshot, err error) {

	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	fsmSnapshot = NewFSMSnapshot(f.state)

	return fsmSnapshot, nil
}

func (f *FSM) Restore(serialized io.ReadCloser) (err error) {

	data, err := ioutil.ReadAll(serialized)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, f.state)
	if err != nil {
		return err
	}

	return nil
}

type FSMSnapshot struct {
	state map[string]interface{}
}

func NewFSMSnapshot(state map[string]interface{}) (s *FSMSnapshot) {

	s = &FSMSnapshot{
		state: make(map[string]interface{}),
	}

	s.state = state

	return s
}

func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) (err error) {

	data, err := json.Marshal(s.state)
	if err != nil {
		return err
	}

	_, err = sink.Write(data)
	if err != nil {
		sink.Cancel()
		return err
	}

	err = sink.Close()
	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (s *FSMSnapshot) Release() {

}
