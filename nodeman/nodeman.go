package nodeman

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/dataptive/styx/logger"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

const (
	dirPerm = 0744
	maxPool = 3
	timeout = 10 * time.Second
	retain  = 1
)

type NodeManager struct {
	config        Config
	boltStore     *raftboltdb.BoltStore
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     *raft.NetworkTransport
	fsm           *FSM
	raftNode      *raft.Raft
}

func NewNodeManager(config Config) (nm *NodeManager, err error) {

	logger.Debugf("nodeman: starting node manager (node_name=%s, state_directory=%s, raft_address=%s)", config.NodeName, config.StateDirectory, config.RaftAddress)

	_, err = os.Stat(config.StateDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		err = os.Mkdir(config.StateDirectory, os.FileMode(dirPerm))
		if err != nil {
			return nil, err
		}
	}

	raftAddr, err := net.ResolveTCPAddr("tcp", config.RaftAddress)
	if err != nil {
		return nil, err
	}

	hcLogger := newHCLogger()

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.NodeName)
	raftConfig.Logger = hcLogger
	raftConfig.ShutdownOnRemove = false

	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(config.StateDirectory, "raft.db"))
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(config.StateDirectory, retain, hcLogger)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransportWithLogger(raftAddr.String(), raftAddr, maxPool, timeout, hcLogger)
	if err != nil {
		return nil, err
	}

	fsm := newFSM()

	raftNode, err := raft.NewRaft(raftConfig, fsm, boltStore, boltStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	nm = &NodeManager{
		config:        config,
		boltStore:     boltStore,
		logStore:      boltStore,
		stableStore:   boltStore,
		snapshotStore: snapshotStore,
		transport:     transport,
		fsm:           fsm,
		raftNode:      raftNode,
	}

	return nm, nil
}

func (nm *NodeManager) Close() (err error) {

	logger.Debug("nodeman: closing node manager")

	future := nm.raftNode.Shutdown()
	err = future.Error()
	if err != nil {
		return err
	}

	err = nm.transport.Close()
	if err != nil {
		return err
	}

	err = nm.boltStore.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(nm.config.StateDirectory)
	if err != nil {
		return err
	}

	return nil
}

func (nm *NodeManager) Bootstrap() (err error) {

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nm.config.NodeName),
				Address: nm.transport.LocalAddr(),
			},
		},
	}

	future := nm.raftNode.BootstrapCluster(configuration)

	err = future.Error()
	if err != nil {
		return err
	}

	return nil
}

func (nm *NodeManager) ListNodes() (nodes []Node, err error) {

	future := nm.raftNode.GetConfiguration()

	err = future.Error()
	if err != nil {
		return nil, err
	}

	configuration := future.Configuration()

	leader := nm.raftNode.Leader()

	for _, server := range configuration.Servers {

		var state raft.RaftState

		if server.Address == leader {
			state = raft.Leader
		} else {
			state = raft.Follower
		}

		node := Node{
			Name:     server.ID,
			State:    state,
			Suffrage: server.Suffrage,
			Address:  server.Address,
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
