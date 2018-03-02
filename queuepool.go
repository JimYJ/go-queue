package queue

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"
)

//Func 队列函数类型
type Func func(value ...interface{})

var (
	defaultTimeOut = 3 * time.Second
	once           sync.Once
	//JobQueue 任务通道
	JobQueue chan *Job
	//QueuePool 队列池
	QueuePool  *queuePool
	finishLock bool
	wg         sync.WaitGroup
	debug      = false
)

type worker struct {
	ID      int           //队列池通道ID
	job     chan *Job     //并发任务结构体
	timeOut time.Duration //超时时间，可自定义
	quit    chan bool
}

//Job 任务结构
type Job struct {
	ID        int64         //任务ID
	FuncQueue Func          //任务函数
	Payload   []interface{} //任务参数
}

type queuePool struct {
	workerChan chan *worker //队列池
}

//InitQueue 初始化队列
func InitQueue(maxConcurrent int, waitLock bool) {
	once.Do(func() {
		QueuePool = &queuePool{
			workerChan: make(chan *worker, maxConcurrent),
		}
		JobQueue = make(chan *Job, maxConcurrent)
		for i := 0; i < maxConcurrent; i++ {
			worker := &worker{
				ID:      i,
				job:     make(chan *Job),
				timeOut: defaultTimeOut,
				quit:    make(chan bool),
			}
			worker.start()
			showLog("worker %d started", worker.ID)
		}
		finishLock = waitLock
		dispatch()
	})
}

func (w *worker) start() {
	go func() {
		QueuePool.workerChan <- w
		id := make(chan int, 1)
		var ctx context.Context
		for {
			select {
			case job := <-w.job:
				showLog("worker: %d, will handle job: %d", w.ID, (*job).ID)
				go w.handleJob(ctx, job, id)
			}
		}
	}()
}

func (w *worker) handleJob(ctx context.Context, job *Job, id chan int) {
	if finishLock {
		wg.Add(1)
		defer wg.Done()
	}
	ctx, _ = context.WithTimeout(context.Background(), w.timeOut)
	queuefunc := (*job).FuncQueue
	value := (*job).Payload
	go func() {
		queuefunc((*job).ID, value)
		select {
		case <-ctx.Done():
			runtime.Goexit()
		default:
			w.quit <- false
		}
	}()
	select {
	case quit := <-w.quit:
		QueuePool.workerChan <- w
		showLog("quit:%t", quit)
		runtime.Goexit()
	case <-ctx.Done():
		QueuePool.workerChan <- w
		showLog("job: %d in woker: %d is timeout...", (*job).ID, w.ID)
		runtime.Goexit()
	}
}

//dispatch 监听JobQueue获取任务
func dispatch() {
	go func() {
		for {
			select {
			case job := <-JobQueue:
				showLog("trying to dispatch job %d ...", (*job).ID)
				worker := <-QueuePool.workerChan
				worker.job <- job
				showLog("job %d dispatched successfully", (*job).ID)

			}
		}
	}()
}

func showLog(str string, value ...interface{}) {
	if debug {
		log.Printf(str, value...)
	}
}

//Done 监听队列是否执行结束
func Done() {
	if finishLock {
		wg.Wait()
	}
}

//Debug 打开日志
func Debug() {
	debug = true
}
