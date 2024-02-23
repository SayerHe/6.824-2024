(cd ../../mrapps && go build $RACE -buildmode=plugin ../../mrapps/crash.so) || exit 1
go run ../mrworker.go ../../mrapps/crash.so