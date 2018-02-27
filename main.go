package main

import (
	"errors"
	"fmt"
	"go-queue/queue"
	"log"
	"runtime"
	"time"
	// "go-queue/queue"
)

type jobFunc func(...interface{}) error

var (
	maps []jobFunc
)

func jobFuncs(a ...interface{}) error {
	fmt.Println(a)
	return errors.New("type")
}

func test2(funcs jobFunc, a, b int) {
	maps = append(maps, funcs)
	funcs(a, b)
}

func test3() {
	a := 1
	b := 2
	// var maps []jobFunc
	// maps = append(maps, test1)
	for i := 1; i < 10; i++ {
		a += i
		b += i
		test2(jobFuncs, a, b)
	}
	fmt.Println(maps[0](a, b))
}

func test4(value ...interface{}) {
	log.Println(value...)
	time.Sleep(5 * time.Second)
}

func main() {
	queue.InitQueue(10)
	for i := 0; i < 1000; i++ {
		job := new(queue.Job)
		job.ID = int64(i)
		job.FuncQueue = test4
		job.Payload = []interface{}{100, 50}
		queue.JobQueue <- job
		log.Println("协程数：", runtime.NumGoroutine())
	}
	log.Println("最终协程数：", runtime.NumGoroutine())
}
