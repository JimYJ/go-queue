# go-queue
go-queue is common queue for concurrency(go-queue 是一个通用的并发通道，可以自定义同时并发数和自定义并发任务)

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
job := new(queue.Job)
job.ID = int64(i)
job.FuncQueue = youfunc
job.Payload = []interface{}{100, 50}
queue.JobQueue <- job
queue.Done()
}
func youfunc(value ...interface){ //you func must meet type Func func(value ...interface{})
    //dosomething
}
```
