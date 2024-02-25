package mr

import (
	"io"
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.Default()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	f, err := os.OpenFile("../../../log/mr.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	defer f.Close()
	multiWriter := io.MultiWriter(os.Stdout, f)
	logger.SetOutput(multiWriter)
	if err != nil {
		logger.Println("fail to open log file")
	}
}
