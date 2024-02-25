package mr

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func invokeReduce(reducef func(string, []string) string, resp *HeartBeatReply) error {
	files := resp.FileName
	//logger.Println(files)
	kvs := make([]KeyValue, 0)
	for _, fileName := range files {
		file, _ := os.Open(fileName)
		decoder := json.NewDecoder(file)
		for {
			kv := KeyValue{}
			err := decoder.Decode(&kv)
			if err != nil {
				break
			}
			kvs = append(kvs, kv)
		}
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Key < kvs[j].Key
	})
	//fmt.Println(kvs)
	var lastKey string
	var sameKey []string
	if len(kvs) > 0 {
		lastKey = kvs[0].Key
		sameKey = []string{kvs[0].Value}
	}
	outputFile, _ := os.CreateTemp("", "mr-out")
	outputFile.Truncate(0)
	for i := 1; i < len(kvs); i++ {
		if kvs[i].Key == lastKey {
			sameKey = append(sameKey, kvs[i].Value)
		} else {
			reduceVal := reducef(lastKey, sameKey)
			fmt.Fprintf(outputFile, "%v %v\n", lastKey, reduceVal)
			lastKey = kvs[i].Key
			sameKey = []string{kvs[i].Value}
		}
		if i == len(kvs)-1 {
			reduceVal := reducef(lastKey, sameKey)
			fmt.Fprintf(outputFile, "%v %v", lastKey, reduceVal)
		}
	}
	fmt.Fprintf(outputFile, "\n")
	os.Rename(outputFile.Name(), fmt.Sprintf("%smr-out-%d", mrConf.resultPath, resp.TaskId))
	reportArgs := &ReportReduceArgs{
		Cookie:       resp.Cookie,
		ReduceTaskId: resp.TaskId,
	}
	reply := ReportReduce(reportArgs)
	if reply.StatusCode != 0 {
		return fmt.Errorf("[REDUCER] Fail to report reducer")
	}
	return nil
}
