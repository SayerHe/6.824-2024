package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	hash    map[string]string
	history map[string]string
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	reply.Value = kv.hash[args.Key]
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.history[args.UUId]; !ok {
		kv.hash[args.Key] = args.Value
		kv.history[args.UUId] = ""
	}
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	if val, ok := kv.history[args.UUId]; !ok {
		reply.Value = kv.hash[args.Key]
		kv.hash[args.Key] += args.Value
		kv.history[args.UUId] = reply.Value
	} else {
		reply.Value = val
	}
	kv.mu.Unlock()
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.hash = make(map[string]string)
	kv.history = make(map[string]string)
	return kv
}
