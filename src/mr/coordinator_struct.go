package mr

import "sync"

type mapTask struct {
	fileNames []string
}

type reduceTask struct {
	fileNames []string
}

type task struct {
	id       int
	taskType mrProcess
	//process           taskProcess
	mapTaskContent    mapTask
	reduceTaskContent reduceTask
}

type worker struct {
	// 无有用信息，对本实现无用
	// TODO：留作后续拓展服务端 PUSH 消息至 worker
	cookie int
	t      *task
}

type stageChecker struct {
	stageCh  chan int
	returnCh chan int // stageChecker 处理完成后，heartBeat协程返回
	tag      mrProcess
}

type Coordinator struct {
	// 假设有很多 mapTask，需要一个专门的slice存储idleMapTasks
	// 假设 mapTask 比较少，可以只用一个 slice 存所有 mapTasks，每次获取空闲任务时都遍历这个slice
	// 对于已经正在处理中的 mapTask，采用hash缓存
	// 当任务完成、失败时能快速更新该 mapTask 状态
	mapIdleTasks          []*task
	mapProcessingTasks    map[int]*task
	mapDoneTasks          map[int]*task
	reduceIdleTasks       []*task
	reduceProcessingTasks map[int]*task
	reduceDoneTasks       map[int]*task
	workers               []*worker
	mu                    sync.Mutex
	nReduce               int
	stage                 stageChecker
}
