package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

// KVStore represents a key-value store
type KVStore struct {
	store map[string]string
}

// Get retrieves the value associated with a key
func (k *KVStore) Get(key string, reply *string) error {
	val, ok := k.store[key]
	if !ok {
		*reply = ""
		return nil
	}
	*reply = val
	return nil
}

// Set associates a value with a key
func (k *KVStore) Set(args [2]string, reply *bool) error {
	k.store[args[0]] = args[1]
	*reply = true
	return nil
}

// RaftFSM represents the Raft state machine for the key-value store
type RaftFSM struct {
	kv    *KVStore
	log   raft.LogStore
	st    raft.StableStore
	trans raft.Transport
}

// Apply applies a Raft log entry to the key-value store
func (fsm *RaftFSM) Apply(logEntry *raft.Log) interface{} {
	if logEntry.Type != raft.LogCommand {
		return nil
	}

	cmd := logEntry.Data
	args := cmd.([2]string)

	var reply bool
	fsm.kv.Set(args, &reply)

	return nil
}

// Snapshot returns a snapshot of the Raft state machine
func (fsm *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &KVStoreSnapshot{store: fsm.kv.store}, nil
}

// Restore restores a snapshot of the Raft state machine
func (fsm *RaftFSM) Restore(snapshot io.ReadCloser) error {
	var store map[string]string
	if err := gob.NewDecoder(snapshot).Decode(&store); err != nil {
		return err
	}

	fsm.kv.store = store

	return nil
}

// KVStoreSnapshot represents a snapshot of the key-value store
type KVStoreSnapshot struct {
	store map[string]string
}

// Persist writes the snapshot to a writer
func (s *KVStoreSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := gob.NewEncoder(sink).Encode(s.store); err != nil {
		sink.Cancel()
		return err
	}

	if err := sink.Close(); err != nil {
		return err
	}

	return nil
}

// Release releases the snapshot
func (s *KVStoreSnapshot) Release() {}

func main() {
	kv := KVStore{store: make(map[string]string)}

	// Configure the Raft server
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node1")

	// Create the log store and stable store
	logStore, err := raftboltdb.NewBoltStore("/tmp/raft/log.db")
	if err != nil {
		log.Fatal(err)
	}
	stableStore, err := raftboltdb.NewBoltStore("/tmp/raft/stable.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create the Raft transport
	trans, err := raft.NewTCPTransport(":8080", nil, 3, time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create the Raft FSM
	fsm := &RaftFSM{kv: &kv, log: logStore, st: stableStore, trans: trans}

	// Create the Raft node
	raftNode, err := raft.NewRaft(config, fsm, logStore, stableStore, nil, trans)
	if err != nil {
		log.Fatal(err)
	}

	// Add the Raft node as a RPC server
	rpcServer := rpc.NewServer()
	rpcServer.Register(&RPCServer{raft: raftNode})
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}

	go rpcServer.Accept(l)

	// Wait forever
	select {}
}

// RPCServer represents the RPC server for the Raft node
type RPCServer struct {
	raft *raft.Raft
}

// Set sets a value in the key-value store
func (s *RPCServer) Set(args [2]string, reply *bool) error {
	cmd := args

	f := s.raft.Apply(cmd, 10*time.Second)
	if err := f.Error(); err != nil {
		return err
	}

	*reply = true

	return nil
}

// Get gets a value from the key-value store
func (s *RPCServer) Get(key string, reply *string) error {
	f := s.raft.Apply([]byte(key), 10*time.Second)
	if err := f.Error(); err != nil {
		return err
	}

	if f.Response() == nil {
		*reply = ""
		return nil
	}

	val := f.Response().(string)
	*reply = val

	return nil
}
