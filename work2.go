package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

// TaskQueue represents a distributed task queue
type TaskQueue struct {
	queue []int
	mu    sync.Mutex
}

// AddTask adds a task to the queue
func (tq *TaskQueue) AddTask(task int, reply *bool) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tq.queue = append(tq.queue, task)

	*reply = true

	return nil
}

// GetTask gets a task from the queue
func (tq *TaskQueue) GetTask(workerID int, reply *int) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.queue) == 0 {
		*reply = -1
		return nil
	}

	task := tq.queue[0]
	tq.queue = tq.queue[1:]

	fmt.Printf("Worker %d got task %d\n", workerID, task)

	*reply = task

	return nil
}

// Master represents a master node in the distributed task queue
type Master struct {
	tq      *TaskQueue
	workers []string
	mu      sync.Mutex
}

// RegisterWorker registers a worker with the master
func (m *Master) RegisterWorker(workerAddr string, reply *int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workers = append(m.workers, workerAddr)

	fmt.Printf("Worker %s registered\n", workerAddr)

	*reply = len(m.workers)

	return nil
}

// Run starts the master node
func (m *Master) Run() {
	rpc.Register(m)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	addr := listener.Addr().String()

	fmt.Printf("Master listening on %s\n", addr)

	go rpc.Accept(listener)

	for {
		// Do nothing, just keep running
	}
}

// Worker represents a worker node in the distributed task queue
type Worker struct {
	masterAddr string
	id         int
}

// Run starts the worker node
func (w *Worker) Run() {
	var reply int

	client, err := jsonrpc.Dial("tcp", w.masterAddr)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Call("Master.RegisterWorker", w.masterAddr, &reply)
	if err != nil {
		log.Fatal(err)
	}

	w.id = reply

	for {
		var task int

		err = client.Call("TaskQueue.GetTask", w.id, &task)
		if err != nil {
			log.Fatal(err)
		}

		if task == -1 {
			fmt.Printf("Worker %d found no tasks, waiting...\n", w.id)
			continue
		}

		fmt.Printf("Worker %d processing task %d\n", w.id, task)
	}
}

func main() {
	tq := TaskQueue{}
	master := Master{tq: &tq}

	go master.Run()

	worker1 := Worker{masterAddr: "localhost:1234"}
	worker2 := Worker{masterAddr: "localhost:1234"}

	go worker1.Run()
	go worker2.Run()

	select {}
}
