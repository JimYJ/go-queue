package queue

import (
	"log"
	"sync"
	"time"
)

//Func 队列函数类型
type Func func(value ...interface{})

var (
	defaultTimeOut = 3 * time.Second
	// HighPriority   = 1
	// MiddlePriority = 2
	// LowPriority    = 3
	once sync.Once
	//JobQueue 任务通道
	JobQueue chan *Job
	//QueuePool 队列池
	QueuePool *queuePool
)

type worker struct {
	ID      int       //队列池通道ID
	job     chan *Job //并发任务结构体
	timeOut time.Duration
	quit    chan bool
}

//Job 任务
type Job struct {
	ID        int64
	FuncQueue Func
	Payload   []interface{}
}

type queuePool struct {
	workerChan chan *worker
}

//InitQueue 初始化队列
func InitQueue(maxConcurrent int) {
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
			log.Printf("worker %d started", worker.ID)
		}
		dispatch()
	})
}

func (w *worker) start() {
	go func() {
		QueuePool.workerChan <- w
		for {
			// t := time.NewTimer(w.timeOut)
			select {
			case job := <-w.job:
				log.Printf("worker: %d, will handle job: %d", w.ID, (*job).ID)
				go w.handleJob(job)
			case quit := <-w.quit:
				QueuePool.workerChan <- w
				log.Println("quit:", quit)
				if quit {
					log.Printf("worker: %d, will stop.", w.ID)
					return
				}
			}
			// select {
			// case <-t.C:
			// 	QueuePool.workerChan <- w
			// 	log.Printf("woker: %d is timeout...", w.ID)
			// }
			// t.Stop()
		}
	}()
}

func (w *worker) handleJob(job *Job) {
	go func() {
		time.Sleep(3 * time.Second)
		log.Printf("job: %d in woker: %d is timeout...", (*job).ID, w.ID)
		// QueuePool.workerChan <- w
		w.quit <- true
	}()
	queuefunc := (*job).FuncQueue
	value := (*job).Payload
	queuefunc((*job).ID, value)
	log.Printf("JobID:%d", (*job).ID)
	w.quit <- false
}

//Dispatch 监听JobQueue获取任务
func dispatch() {
	go func() {
		for {
			select {
			case job := <-JobQueue:
				go func(job *Job) {
					// log.Printf("trying to dispatch job %d ...", (*job).ID)
					worker := <-QueuePool.workerChan
					worker.job <- job
					// log.Printf("job %d dispatched successfully", (*job).ID)
				}(job)
			}
		}
	}()
}
