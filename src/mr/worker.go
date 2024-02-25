package mr

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// Worker main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	args := &HeartBeatArgs{
		Cookie: -1,
	}
	for {
		resp := HeartBeat(args)
		if resp.StatusCode == -1 {
			time.Sleep(time.Second * 5)
			continue
		}
		switch resp.TaskType {
		case Map_Task:
			invokeMap(mapf, resp)
		case Reduce_Task:
			invokeReduce(reducef, resp)
		case Wait_Task:
			time.Sleep(time.Second * 1)
		case Done_Task:
			return
		default:
			return
		}
	}
}

func HeartBeat(args *HeartBeatArgs) *HeartBeatReply {
	reply := &HeartBeatReply{}
	ok := call("Coordinator.HeartBeat", args, reply)
	if !ok {
		fmt.Printf("call hearBeat failed!\n")
	}
	// update cookie
	args.Cookie = reply.Cookie
	//fmt.Printf("RESP: %v\n", reply)
	return reply
}

func ReportMap(args *ReportMapArgs) *ReportMapReply {
	reply := ReportMapReply{}
	ok := call("Coordinator.ReportMap", args, &reply)
	if !ok {
		fmt.Printf("call report failed!\n")
	}
	return &reply
}

func ReportReduce(args *ReportReduceArgs) *ReportReduceReply {
	reply := ReportReduceReply{}
	ok := call("Coordinator.ReportReduce", args, &reply)
	if !ok {
		fmt.Printf("call report failed!\n")
	}
	return &reply
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	//c, err := rpc.DialHTTP("tcp", "localhost:1500")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
