package mr

import (
	"6.5840/utils"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

func (c *Coordinator) loadSourceFile(root, prefix string) error {
	files, err := utils.FindFilesStartWith(root, prefix)
	if err != nil {
		return err
	}
	for id, file := range files {
		c.mapIdleTasks = append(c.mapIdleTasks,
			&task{
				id:       id,
				taskType: MapProcess,
				mapTaskContent: mapTask{
					fileNames: []string{file},
				},
			})
	}
	//log.Println(c.mapIdleTasks[0].mapTaskContent)
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.HandleHTTP()
	err := rpc.Register(c)
	if err != nil {
		return
	}
	err = c.loadSourceFile(mrConf.root, mrConf.prefix)
	if err != nil {
		return
	}
	//l, e := net.Listen("tcp", "localhost:1500")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
	go c.checkStage()
	//log.Println("server start on localhost:1234")
}

// Done main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	done := c.stage.tag == DoneProcess
	c.mu.Unlock()
	return done
}

// MakeCoordinator create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		mapIdleTasks:          make([]*task, 0),
		mapProcessingTasks:    make(map[int]*task),
		mapDoneTasks:          make(map[int]*task),
		workers:               make([]*worker, 0),
		reduceIdleTasks:       make([]*task, 0),
		reduceProcessingTasks: make(map[int]*task),
		reduceDoneTasks:       make(map[int]*task),
		nReduce:               nReduce,
		stage: stageChecker{
			stageCh:  make(chan int),
			returnCh: make(chan int),
			tag:      MapProcess,
		},
	}
	for i := 0; i < nReduce; i++ {
		c.reduceIdleTasks = append(c.reduceIdleTasks, &task{
			id:       i,
			taskType: ReduceProcess,
			reduceTaskContent: reduceTask{
				fileNames: make([]string, 0),
			},
		})
	}
	c.server()
	return &c
}
