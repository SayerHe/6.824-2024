package mr

import "time"

func (c *Coordinator) HeartBeat(args *HeartBeatArgs, reply *HeartBeatReply) error {
	w := c.getWorker(args.Cookie)
	//logger.Printf("receive heartbeat from %d", w.cookie)
	// 读 mr 任务进度 + 具体任务处理应该是一个原子行为
	// 即 读取到一个 mr 任务进度后，必须要完整执行该状态下对应的 handler；
	// 否则一些 worker 会读到错的 mr 进度
	// 比如可能出现任务队列只剩一个任务，两个 worker 同时读到 MapProcess，导致panic
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.stage.tag {
	case MapProcess:
		c.handleMap(w, reply)
	case ReduceProcess:
		c.handleReduce(w, reply)
	case WaitForDoneProcess:
		c.handleWait(w, reply)
	case WaitForReduceProcess:
		c.handleWait(w, reply)
	case DoneProcess:
		c.handleDone(w, reply)
	}
	reply.NReduce = c.nReduce
	return nil
}

func (c *Coordinator) asyncTaskTimer(retryTime time.Duration, taskId int, flag int) {
	select {
	case <-time.After(retryTime):
		c.mu.Lock()
		if flag == 0 {
			// 探测 map 任务是否已经完成
			if _, ok := c.mapDoneTasks[taskId]; !ok {
				c.mapIdleTasks = append(c.mapIdleTasks, c.mapProcessingTasks[taskId])
				//logger.Printf("map task %d retry\n", taskId)
				c.stage.stageCh <- MapRetry_Sig
				<-c.stage.returnCh
			}
		} else if flag == 1 {
			// 探测 reduce 是否已经完成
			if _, ok := c.reduceDoneTasks[taskId]; !ok {
				c.reduceIdleTasks = append(c.reduceIdleTasks, c.reduceProcessingTasks[taskId])
				//logger.Printf("reduce task %d retry\n", taskId)
				c.stage.stageCh <- ReduceRetry_Sig
				<-c.stage.returnCh
			}
		}
		c.mu.Unlock()
	}
}

func (c *Coordinator) handleMap(w *worker, reply *HeartBeatReply) {
	w.t = c.mapIdleTasks[0]
	reply.TaskType = Map_Task
	reply.Cookie = w.cookie
	releaseMapTask := c.mapIdleTasks[0]
	reply.FileName = releaseMapTask.mapTaskContent.fileNames
	reply.TaskId = releaseMapTask.id
	c.mapProcessingTasks[reply.TaskId] = releaseMapTask
	c.mapIdleTasks = c.mapIdleTasks[1:]
	c.stage.stageCh <- MapRelease_Sig
	<-c.stage.returnCh
	go c.asyncTaskTimer(10*time.Second, releaseMapTask.id, 0)
	//logger.Printf("worker %d get map task %d\n", w.cookie, reply.TaskId)
}

func (c *Coordinator) handleReduce(w *worker, reply *HeartBeatReply) {
	releaseReduceTask := c.reduceIdleTasks[0]
	w.t = releaseReduceTask
	reply.TaskType = Reduce_Task
	reply.Cookie = w.cookie
	reply.FileName = releaseReduceTask.reduceTaskContent.fileNames
	reply.TaskId = releaseReduceTask.id
	c.reduceProcessingTasks[reply.TaskId] = releaseReduceTask
	c.reduceIdleTasks = c.reduceIdleTasks[1:]
	c.stage.stageCh <- ReduceRelease_Sig
	<-c.stage.returnCh
	go c.asyncTaskTimer(10*time.Second, releaseReduceTask.id, 1)
	//logger.Printf("worker %d get reduce task %d\n", w.cookie, reply.TaskId)
}

func (c *Coordinator) handleWait(w *worker, reply *HeartBeatReply) {
	//logger.Printf("worker %d call handleWait\n", w.cookie)
	reply.TaskType = Wait_Task
	reply.Cookie = w.cookie
}

func (c *Coordinator) handleDone(w *worker, reply *HeartBeatReply) {
	//logger.Printf("worker %d call handleDone\n", w.cookie)
	reply.TaskType = Done_Task
	reply.Cookie = w.cookie
}

func (c *Coordinator) getWorker(cookie int) *worker {
	var w *worker
	if cookie == -1 {
		c.mu.Lock()
		w = &worker{
			cookie: len(c.workers),
		}
		c.workers = append(c.workers, w)
		c.mu.Unlock()
	} else {
		// only idle mapper can send heartbeat the coordinator
		c.mu.Lock()
		w = c.workers[cookie]
		c.mu.Unlock()
	}
	return w
}

func (c *Coordinator) checkStage() {
	c.mu.Lock()
	mapTaskCnt := len(c.mapIdleTasks)
	reduceTaskCnt := c.nReduce
	c.mu.Unlock()
	mapTaskRelease, mapTaskDone, reduceTaskRelease, reduceTaskDone := 0, 0, 0, 0
	for {
		switch <-c.stage.stageCh {
		case MapDone_Sig:
			mapTaskDone++
		case ReduceDone_Sig:
			reduceTaskDone++
		case MapRelease_Sig:
			mapTaskRelease++
		case ReduceRelease_Sig:
			reduceTaskRelease++
		case MapRetry_Sig:
			mapTaskRelease--
		case ReduceRetry_Sig:
			reduceTaskRelease--
		}
		//logger.Println(mapTaskRelease, mapTaskDone, reduceTaskRelease, reduceTaskDone)
		if c.stage.tag == MapProcess && mapTaskRelease == mapTaskCnt {
			c.stage.tag = WaitForReduceProcess
			//logger.Printf("状态转为 wait for reduce")
		} else if c.stage.tag == WaitForReduceProcess && mapTaskDone == mapTaskCnt {
			c.stage.tag = ReduceProcess
			//logger.Printf("状态转为 reduce")
		} else if c.stage.tag == ReduceProcess && reduceTaskRelease == reduceTaskCnt {
			c.stage.tag = WaitForDoneProcess
			//logger.Printf("状态转为 wait for done")
		} else if c.stage.tag == WaitForDoneProcess && reduceTaskDone == reduceTaskCnt {
			c.stage.tag = DoneProcess
			//logger.Printf("状态转为 done")
		} else if c.stage.tag == WaitForReduceProcess && mapTaskRelease < mapTaskCnt {
			c.stage.tag = MapProcess
			//logger.Printf("出现重试 状态转为 map")
		} else if c.stage.tag == WaitForDoneProcess && reduceTaskRelease < reduceTaskCnt {
			c.stage.tag = ReduceProcess
			//logger.Printf("出现重试 状态转为 reduce")
		}
		// 确保每个状态信号被完全处理，且状态转化完成后，才释放资源给其它 worker
		c.stage.returnCh <- 1
	}
}
