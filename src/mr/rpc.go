package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type HeartBeatArgs struct {
	Cookie int
}

type HeartBeatReply struct {
	FileName   []string
	NReduce    int
	TaskId     int
	TaskType   int
	Cookie     int
	StatusCode int
}

type ReportMapArgs struct {
	Cookie            int
	MapTaskId         int
	IntermediateFiles map[int]string
}

type ReportMapReply struct {
	StatusCode int
}

type ReportReduceArgs struct {
	Cookie       int
	ReduceTaskId int
}

type ReportReduceReply struct {
	StatusCode int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
