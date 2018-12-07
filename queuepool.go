package queue

import (
	"log"
	"runtime"
	"sync"
	"time"
)

//Func 队列函数类型
type Func func(value ...interface{}) error

var (
	defaultTimeOut = 3 * time.Second
	once           sync.Once
	//JobQueue 任务通道
	JobQueue chan *Job
	//QueuePool 队列池
	QueuePool  *queuePool
	retryQueue chan *Job
	//FailList 失败列表
	FailList           = make([]*Job, 0)
	finishLock         bool
	wg                 sync.WaitGroup
	concurrentInterval time.Duration
	useInterval        = false
	debug              = false
	retry              = false
	maxRetryTimes      = 3
	retryInterval      = 3 * time.Second
)

type worker struct {
	ID      int           //队列池通道ID
	job     chan *Job     //并发任务结构体
	timeOut time.Duration //超时时间，可自定义
	quit    chan bool
}

//Job 任务结构
type Job struct {
	ID         int64         //任务ID
	FuncQueue  Func          //任务函数
	Payload    []interface{} //任务参数
	RetryTimes int           //重试次数
}

type queuePool struct {
	workerChan chan *worker //队列池
}

//InitQueue 初始化队列
func InitQueue(maxConcurrent int, waitLock bool, useRetry bool) {
	once.Do(func() {
		QueuePool = &queuePool{
			workerChan: make(chan *worker, maxConcurrent),
		}
		JobQueue = make(chan *Job, maxConcurrent)
		if useRetry {
			retryQueue = make(chan *Job, maxConcurrent)
		}
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
		retry = useRetry
		dispatch()
	})
}

func (w *worker) start() {
	go func() {
		QueuePool.workerChan <- w
		id := make(chan int, 1)
		timeout := time.NewTimer(w.timeOut)
		for {
			timeout.Reset(w.timeOut)
			select {
			case job := <-w.job:
				showLog("worker: %d, will handle job: %d", w.ID, (*job).ID)
				w.handleJob(timeout, job, id)
			}
		}
	}()
}

func (w *worker) handleJob(timeout *time.Timer, job *Job, id chan int) {
	if finishLock {
		wg.Add(1)
		defer wg.Done()
	}
	queuefunc := (*job).FuncQueue
	value := (*job).Payload
	go func() {
		err := queuefunc(value...)
		select {
		case <-timeout.C:
			runtime.Goexit()
		default:
			if err != nil {
				w.quit <- true
			} else {
				w.quit <- false
			}
			runtime.Goexit()
		}
	}()
	select {
	case quit := <-w.quit:
		waitInterval()
		go retryHandle(quit, job)
		QueuePool.workerChan <- w
		showLog("quit:%t", quit)
		// runtime.Goexit()
	case <-timeout.C:
		waitInterval()
		go retryHandle(true, job)
		QueuePool.workerChan <- w
		showLog("job: %d in woker: %d is timeout...", (*job).ID, w.ID)
		// runtime.Goexit()
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
			case job := <-retryQueue:
				showLog("retrying to dispatch job %d for %d times ...", (*job).ID, (*job).RetryTimes)
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

//SetConcurrentInterval 设置并发时间间隔
func SetConcurrentInterval(interval time.Duration) {
	if interval > 0 {
		concurrentInterval = interval
		useInterval = true
	}
}

//Debug 打开日志
func Debug() {
	debug = true
}

func waitInterval() {
	if useInterval {
		time.Sleep(concurrentInterval)
	}
}

//Push 推送任务
func Push(job *Job) {
	JobQueue <- job
}

func retryHandle(quit bool, job *Job) {
	if finishLock {
		wg.Add(1)
		defer wg.Done()
	}
	if quit && retry {
		(*job).RetryTimes++
		if (*job).RetryTimes <= maxRetryTimes {
			time.Sleep(retryInterval)
			retryQueue <- job
		} else {
			FailList = append(FailList, job)
		}
	} else if quit && !retry {
		FailList = append(FailList, job)
	}
}
