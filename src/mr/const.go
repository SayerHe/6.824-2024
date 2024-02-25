package mr

type mrProcess = int

const (
	MapProcess           mrProcess = 1
	ReduceProcess        mrProcess = 2
	WaitForReduceProcess mrProcess = 3
	WaitForDoneProcess   mrProcess = 4
	DoneProcess          mrProcess = 5
)

const (
	Map_Task    = 1
	Wait_Task   = 2
	Reduce_Task = 3
	Done_Task   = 4
)

const (
	MapDone_Sig       = 1
	ReduceDone_Sig    = 2
	MapRelease_Sig    = 3
	ReduceRelease_Sig = 4
	MapRetry_Sig      = 5
	ReduceRetry_Sig   = 6
)
