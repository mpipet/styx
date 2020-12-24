//
// TODO:
// - Déplacement des fonctions d'upgrade dans un package commun (api ?)
// - sur les opérations qui requierent un leader, si il n'y en a pas, attendre avec un timeout
//
package nodeman

import (
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"time"

	"gitlab.com/dataptive/styx/logger"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

const (
	HashicorpRaftProtocolString = "hashicorp-raft/3"
)

const (
	transportMaxPool  = 3
	transportTimeout  = 10 * time.Second
	snapshotsRetained = 1
	applyTimeout = 10 * time.Second

	// See https://godoc.org/github.com/hashicorp/raft#Config
	heartbeatTimeout   = 1000 * time.Millisecond
	electionTimeout    = 1000 * time.Millisecond
	commitTimeout      = 50 * time.Millisecond
	maxAppendEntries   = 64
	trailingLogs       = 10240
	snapshotInterval   = 120 * time.Second
	snapshotThreshold  = 8192
	leaderLeaseTimeout = 500 * time.Millisecond
)

var (
	ErrNotLeader = errors.New("nodeman: not a leader")
)

type Node struct {
	Name     raft.ServerID
	Leader   bool
	Suffrage raft.ServerSuffrage
	Address  raft.ServerAddress
}

type NodeManager struct {
	config        Config
	boltStore     *raftboltdb.BoltStore
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     *raft.NetworkTransport
	raftStream    *streamLayer
	fsm           *FSM
	raftNode      *raft.Raft
}

func NewNodeManager(config Config) (nm *NodeManager, err error) {

	logger.Debugf("nodeman: starting node manager (node_name=%s, raft_directory=%s, advertise_address=%s)", config.NodeName, config.RaftDirectory, config.AdvertiseAddress)

	hcLogger := newHCLogger()

	raftConfig := raft.DefaultConfig()
	raftConfig.HeartbeatTimeout = heartbeatTimeout
	raftConfig.ElectionTimeout = electionTimeout
	raftConfig.CommitTimeout = commitTimeout
	raftConfig.MaxAppendEntries = maxAppendEntries
	raftConfig.TrailingLogs = trailingLogs
	raftConfig.SnapshotInterval = snapshotInterval
	raftConfig.SnapshotThreshold = snapshotThreshold
	raftConfig.LeaderLeaseTimeout = leaderLeaseTimeout

	raftConfig.LocalID = raft.ServerID(config.NodeName)
	raftConfig.Logger = hcLogger

	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(config.RaftDirectory, "raft.db"))
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(config.RaftDirectory, snapshotsRetained, hcLogger)
	if err != nil {
		return nil, err
	}

	advertiseAddr, err := net.ResolveTCPAddr("tcp", config.AdvertiseAddress)
	if err != nil {
		return nil, err
	}

	raftStream := newStreamLayer(advertiseAddr)

	transport := raft.NewNetworkTransport(raftStream, transportMaxPool, transportTimeout, nil)
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
		raftStream:    raftStream,
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

	return nil
}

func (nm *NodeManager) AcceptHandler(conn net.Conn) {

	nm.raftStream.acceptHandler(conn)
}

func (nm *NodeManager) IsLeader() (is bool) {

	return nm.raftNode.State() == raft.Leader
}

func (nm *NodeManager) Leader() (leader raft.ServerAddress) {

	return nm.raftNode.Leader()
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
	leaderAddr := nm.raftNode.Leader()

	for _, server := range configuration.Servers {

		var leader bool

		if server.Address == leaderAddr {
			leader = true
		} else {
			leader = false
		}

		node := Node{
			Name:     server.ID,
			Leader:   leader,
			Suffrage: server.Suffrage,
			Address:  server.Address,
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (nm *NodeManager) AddNode(nodeName raft.ServerID, nodeAddr raft.ServerAddress, voter bool) (err error) {

	if nm.raftNode.State() != raft.Leader {
		return ErrNotLeader
	}

	if voter {
		future := nm.raftNode.AddVoter(nodeName, nodeAddr, 0, 0)
		err = future.Error()
		if err != nil {
			return err
		}
	} else {
		future := nm.raftNode.AddNonvoter(nodeName, nodeAddr, 0, 0)
		err = future.Error()
		if err != nil {
			return err
		}
	}

	return nil
}

func (nm *NodeManager) RemoveNode(nodeName raft.ServerID) (err error) {

	if nm.raftNode.State() != raft.Leader {
		return ErrNotLeader
	}

	future := nm.raftNode.RemoveServer(nodeName, 0, 0)
	err = future.Error()
	if err != nil {
		return err
	}

	return nil
}

func (nm *NodeManager) GetUnsafe(key string) (value interface{}, err error){

	state := nm.fsm.GetState()

	value, ok := state[key]
	if !ok {
		return nil, ErrNotFound
	}

	return value, nil
}

func (nm *NodeManager) Get(key string) (value interface{}, err error){

	if nm.raftNode.State() != raft.Leader {
		return nil, ErrNotLeader
	}

	command := FSMCommand{
		Operation: "get",
		Key: key,
	}

	data, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	future := nm.raftNode.Apply(data, applyTimeout)
	err = future.Error()
	if err != nil {
		return nil, err
	}

	result := future.Response().(*FSMResult)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Value, nil
}

func (nm *NodeManager) Set(key string, value interface{}) (err error) {

	if nm.raftNode.State() != raft.Leader {
		return ErrNotLeader
	}

	command := FSMCommand{
		Operation: "set",
		Key: key,
		Value: value,
	}

	data, err := json.Marshal(command)
	if err != nil {
		return err
	}

	future := nm.raftNode.Apply(data, applyTimeout)
	err = future.Error()
	if err != nil {
		return err
	}

	result := future.Response().(*FSMResult)

	if result.Err != nil {
		return result.Err
	}

	return nil
}

func (nm *NodeManager) Delete(key string) (err error) {

	if nm.raftNode.State() != raft.Leader {
		return ErrNotLeader
	}

	command := FSMCommand{
		Operation: "delete",
		Key: key,
	}

	data, err := json.Marshal(command)
	if err != nil {
		return err
	}

	future := nm.raftNode.Apply(data, applyTimeout)
	err = future.Error()
	if err != nil {
		return err
	}

	result := future.Response().(*FSMResult)

	if result.Err != nil {
		return result.Err
	}

	return nil
}

func (nm *NodeManager) ListUnsafe() (value interface{}, err error){

	state := nm.fsm.GetState()

	return state, nil
}

func (nm *NodeManager) List() (value interface{}, err error){

	if nm.raftNode.State() != raft.Leader {
		return nil, ErrNotLeader
	}

	command := FSMCommand{
		Operation: "list",
	}

	data, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	future := nm.raftNode.Apply(data, applyTimeout)
	err = future.Error()
	if err != nil {
		return nil, err
	}

	result := future.Response().(*FSMResult)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Value, nil
}
