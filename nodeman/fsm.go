package nodeman

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/raft"
)

type FSM struct {
	state     []byte
	stateLock sync.Mutex
}

func newFSM() (fsm *FSM) {

	fsm = &FSM{
		state:     []byte{},
		stateLock: sync.Mutex{},
	}

	return fsm
}

func (f *FSM) GetState() (state []byte) {

	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	state = f.state

	return state
}

func (f *FSM) Apply(log *raft.Log) (applier interface{}) {

	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	f.state = log.Data

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

	f.state = data

	return nil
}

type FSMSnapshot struct {
	state []byte
}

func NewFSMSnapshot(state []byte) (s *FSMSnapshot) {

	s = &FSMSnapshot{
		state: []byte{},
	}

	s.state = state

	return s
}

func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) (err error) {

	_, err = sink.Write(s.state)
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
