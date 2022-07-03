package kvraft

import (
	"crypto/rand"
	"math/big"

	"6.824/labrpc"
)

type Clerk struct {
	servers []*labrpc.ClientEnd
	// You will have to modify this struct.
	me     int64 // my client id
	leader int   // remember which server turned out to be the leader for the last RPC
	opId   int   // operation id, increase monotonically
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	// You'll have to add code here.
	ck.me = nrand()
	ck.leader = 0
	ck.opId = 1
	return ck
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) Get(key string) string {
	// You will have to modify this function.
	args := &GetArgs{
		Key:      key,
		ClientId: ck.me,
		OpId:     ck.opId,
	}
	ck.opId++
	serverId := ck.leader // send to leader first
	for {
		reply := &GetReply{}
		ok := ck.servers[serverId].Call("KVServer.Get", args, reply)
		if !ok || reply.Err == ErrWrongLeader {
			// no reply (reply dropped, network partition, etc.) or
			// wrong leader,
			// try next server
			serverId = (serverId + 1) % len(ck.servers)
			continue
		}

		ck.leader = serverId // remember current leader
		if reply.Err == ErrNoKey {
			return ""
		}
		if reply.Err == OK {
			return reply.Value
		}
	}
}

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
func (ck *Clerk) PutAppend(key string, value string, op opType) {
	// You will have to modify this function.
	args := &PutAppendArgs{
		Key:      key,
		Value:    value,
		Op:       op,
		ClientId: ck.me,
		OpId:     ck.opId,
	}
	ck.opId++
	serverId := ck.leader
	for {
		reply := &PutAppendReply{}
		ok := ck.servers[serverId].Call("KVServer.PutAppend", args, reply)
		if !ok || reply.Err == ErrWrongLeader {
			serverId = (serverId + 1) % len(ck.servers)
			continue
		}

		ck.leader = serverId
		if reply.Err == OK {
			return
		}
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, opPut)
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, opAppend)
}
