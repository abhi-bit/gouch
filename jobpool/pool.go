package pool

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

//Worker to be implemented in order to perform work
type Worker interface {
	DoWork(workRoutine int)
}

type (
	poolWork struct {
		work          Worker
		resultChannel chan error
	}

	//WorkPool implements a job pool
	WorkPool struct {
		shutdownQueueChannel chan string
		shutdownWorkChannel  chan struct{}
		shutdownWaitGroup    sync.WaitGroup
		queueChannel         chan poolWork
		workChannel          chan Worker
		queuedWork           int64
		activeRoutines       int64
		queueCapacity        int64
	}
)

func init() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

//New creates a job pool
func New(numberOfRoutines int, queueCapacity int64) (workPool *WorkPool) {
	workPool = &WorkPool{
		shutdownQueueChannel: make(chan string),
		shutdownWorkChannel:  make(chan struct{}),
		queueChannel:         make(chan poolWork),
		workChannel:          make(chan Worker, queueCapacity),
		queuedWork:           0,
		activeRoutines:       0,
		queueCapacity:        queueCapacity,
	}

	for workRoutine := 0; workRoutine < numberOfRoutines; workRoutine++ {
		workPool.shutdownWaitGroup.Add(1)
		go workPool.workRoutine(workRoutine)
	}

	go workPool.queueRoutine()
	return
}

func (workPool *WorkPool) workRoutine(workRoutine int) {
	for {
		select {
		case <-workPool.shutdownWorkChannel:
			writeStdout(fmt.Sprintf("WorkRoutine %d", workRoutine), "workRoutine", "Going Down")
			workPool.shutdownWaitGroup.Done()
			return

		case poolWorker := <-workPool.workChannel:
			workPool.safelyDoWork(workRoutine, poolWorker)
			break
		}
	}
}

func (workPool *WorkPool) safelyDoWork(workRoutine int, poolWorker Worker) {
	defer catchPanic(nil, "WorkRoutine", "SafelyDoWork")
	defer func() {
		atomic.AddInt64(&workPool.activeRoutines, -1)
	}()

	atomic.AddInt64(&workPool.queuedWork, -1)
	atomic.AddInt64(&workPool.activeRoutines, 1)

	poolWorker.DoWork(workRoutine)
}

func (workPool *WorkPool) queueRoutine() {
	for {
		select {
		case <-workPool.shutdownQueueChannel:
			writeStdout("Queue", "queueRoutine", "Going Down")
			workPool.shutdownQueueChannel <- "Down"
			return

		case queueItem := <-workPool.queueChannel:
			if atomic.AddInt64(&workPool.queuedWork, 0) == workPool.queueCapacity {
				queueItem.resultChannel <- fmt.Errorf("Thread Pool at Capacity")
				continue
			}

			atomic.AddInt64(&workPool.queuedWork, 1)
			workPool.workChannel <- queueItem.work
			queueItem.resultChannel <- nil
			break
		}
	}
}

//PostWork will delegate work in WorkPool
func (workPool *WorkPool) PostWork(goRoutine string, work Worker) (err error) {
	defer catchPanic(&err, goRoutine, "PostWork")

	poolWork := poolWork{work, make(chan error)}

	defer close(poolWork.resultChannel)

	workPool.queueChannel <- poolWork
	err = <-poolWork.resultChannel

	return
}

//Shutdown will release resources
func (workPool *WorkPool) Shutdown(goRoutine string) (err error) {
	defer catchPanic(&err, goRoutine, "Shutdown")

	writeStdout(goRoutine, "Shutdown", "Started")
	writeStdout(goRoutine, "Started", "Queue Routine")

	workPool.shutdownQueueChannel <- "Down"
	<-workPool.shutdownQueueChannel

	close(workPool.queueChannel)
	close(workPool.shutdownQueueChannel)

	writeStdout(goRoutine, "Shutdown", "Shutting Down Work Routines")

	close(workPool.shutdownWorkChannel)
	workPool.shutdownWaitGroup.Wait()

	close(workPool.workChannel)

	writeStdout(goRoutine, "Shutdown", "Completed")

	return
}

func catchPanic(err *error, goRoutine string, functionName string) {
	if r := recover(); r != nil {
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)

		writeStdoutf(goRoutine, functionName, "PANIC Deffered [%v] : Stack Trace : %v", r, string(buf))

		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}

func writeStdout(goRoutine string, functionName string, message string) {
	log.Printf("%s : %s : %s\n", goRoutine, functionName, message)
}

func writeStdoutf(goRoutine string, functionName string, format string, a ...interface{}) {
	writeStdout(goRoutine, functionName, fmt.Sprintf(format, a...))
}

//QueuedWork will return the current count of work items
func (workPool *WorkPool) QueuedWork() int64 {
	return atomic.AddInt64(&workPool.queuedWork, 0)
}

//ActiveRoutines will return number of routines currently performing work
func (workPool *WorkPool) ActiveRoutines() int64 {
	return atomic.AddInt64(&workPool.activeRoutines, 0)
}
