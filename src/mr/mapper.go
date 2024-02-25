package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"strings"
)

type tmpFile struct {
	file    *os.File
	encoder *json.Encoder
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func invokeMap(mapf func(string, string) []KeyValue, resp *HeartBeatReply) error {
	//logger.Printf("RESP in invokeMap: %v \n", resp)
	fileName := resp.FileName[0]
	file, _ := os.Open(fileName)
	var content string
	contentBuffer := make([]byte, 1024)
	for {
		n, err := file.Read(contentBuffer)
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Println("mapper fail to open file: ", fileName)
			return err
		}
		content += string(contentBuffer[:n])

	}
	kvs := mapf(fileName, content)
	//logger.Println(kvs)
	outputFiles, err := writeToIntermediate(kvs, *resp)
	if err != nil {
		return nil
	}
	reportArgs := &ReportMapArgs{
		Cookie:            resp.Cookie,
		MapTaskId:         resp.TaskId,
		IntermediateFiles: outputFiles,
	}
	reply := ReportMap(reportArgs)
	if reply.StatusCode != 0 {
		return fmt.Errorf("[MAPPER] Fail to report mapper result")
	}
	// fmt.Println("[MAPPER] Successfully report")
	return nil
}

func writeToIntermediate(kvs []KeyValue, resp HeartBeatReply) (map[int]string, error) {
	outputFiles := make(map[int]tmpFile)
	intermidateFiles := make(map[int]string)
	for i := 0; i < resp.NReduce; i++ {
		file, _ := os.CreateTemp("", fmt.Sprintf("intermediate-%d-%d-", resp.TaskId, i))
		err := file.Truncate(0)
		if err != nil {
			fmt.Println("[MAPPER] Fail to open intermediate file")
			return nil, err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		outputFiles[i] = tmpFile{
			file:    file,
			encoder: json.NewEncoder(file),
		}
		intermidateFiles[i] = fmt.Sprintf("../%sout-%d-%d", mrConf.intermediatePath, resp.TaskId, i)
	}
	for _, kv := range kvs {
		k := kv.Key
		err := outputFiles[ihash(k)%resp.NReduce].encoder.Encode(kv)
		if err != nil {
			fmt.Printf("[MAPPER] Fail to write intermediate file: %v\n", err)
			return nil, err
		}
	}
	for _, tmp := range outputFiles {
		taskId, reduceId := strings.Split(tmp.file.Name(), "-")[1], strings.Split(tmp.file.Name(), "-")[2]
		os.Rename(tmp.file.Name(), fmt.Sprintf("../%sout-%s-%s", mrConf.intermediatePath, taskId, reduceId))
	}
	return intermidateFiles, nil
}
