package mr

func (c *Coordinator) ReportMap(args *ReportMapArgs, reply *ReportMapReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 任务超时，其它 worker 先完成了该任务
	if _, ok := c.mapDoneTasks[args.MapTaskId]; ok {
		return nil
	}
	// record map result
	for k, v := range args.IntermediateFiles {
		c.reduceIdleTasks[k].reduceTaskContent.fileNames = append(c.reduceIdleTasks[k].reduceTaskContent.fileNames, v)
	}
	c.mapDoneTasks[args.MapTaskId] = c.mapProcessingTasks[args.MapTaskId]
	delete(c.mapProcessingTasks, args.MapTaskId)
	c.stage.stageCh <- MapDone_Sig
	<-c.stage.returnCh
	return nil
}

func (c *Coordinator) ReportReduce(args *ReportReduceArgs, reply *ReportReduceReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 任务超时，coordinator 将任务重新分配，并且其它 worker 先完成了任务
	if _, ok := c.reduceDoneTasks[args.ReduceTaskId]; ok {
		return nil
	}
	c.reduceDoneTasks[args.ReduceTaskId] = c.reduceProcessingTasks[args.ReduceTaskId]
	delete(c.reduceProcessingTasks, args.ReduceTaskId)
	c.stage.stageCh <- ReduceDone_Sig
	<-c.stage.returnCh

	return nil
}
