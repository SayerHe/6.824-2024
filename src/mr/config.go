package mr

type config struct {
	intermediatePath string
	resultPath       string
	root             string
	prefix           string
}

var mrConf config = config{
	intermediatePath: "intermediates/",
	resultPath:       "",
	root:             "/Users/bytedance/code/golang/6.824-2024/src/main/",
	prefix:           "pg-",
}
