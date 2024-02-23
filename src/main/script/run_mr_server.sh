rm mr-out*
go build -buildmode=plugin ../../mrapps/crash.go
go run -race ../mrcoordinator.go pg-*.txt