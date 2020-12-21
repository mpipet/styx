package nodeman

import (
	"github.com/hashicorp/raft"
)

type Node struct {
	Name     raft.ServerID
	State    raft.RaftState
	Suffrage raft.ServerSuffrage
	Address  raft.ServerAddress
}
