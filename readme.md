[![Build Status](https://travis-ci.org/JimYJ/go-queue.svg?branch=master)](https://travis-ci.org/JimYJ/go-queue)
[![Go Report Card](https://goreportcard.com/badge/github.com/JimYJ/go-queue)](https://goreportcard.com/report/github.com/JimYJ/go-queue)

# go-queue
go-queue is task queue for concurrency(go-queue 是一个通用的并发通道，可以自定义同时并发数和自定义并发任务)

## How to get

```
go get github.com/JimYJ/go-queue
```

## Usage

**import:**

```go
import "github.com/JimYJ/go-queue"
```

**init and use:**

```go
func main() {
    queue.InitQueue(10, true)//true means main goroutine will wait that all queue done 
    queue.SetConcurrentInterval(1 * time.Millisecond)// set interval time for each concurrent， default 0
    for i:=0;i<1000;i++{
        job := new(queue.Job)
        job.ID = int64(i)
        job.FuncQueue = youfunc
        job.Payload = []interface{}{100, 50}
        queue.JobQueue <- job
    }
    queue.Done()
}

func youfunc(value ...interface){ //you func must meet type Func func(value ...interface{})
    //dosomething
}
```
