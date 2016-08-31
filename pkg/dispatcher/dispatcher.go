package dispatcher

import (
	"os"
	"strconv"
)

var (
	MaxWorker = os.Getenv("MAX_WORKERS")
	MaxQueue  = os.Getenv("MAX_QUEUE")
)

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	max := 2
	if maxWorkers > 0 {
		max = maxWorkers
	} else {
		if v, err := strconv.Atoi(MaxWorker); err == nil {
			if v > 0 {
				max = v
			}
		}
	}
	pool := make(chan chan Job, max)
	return &Dispatcher{WorkerPool: pool, maxWorkers: max}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}

func (d *Dispatcher) MaxWorkers() int {
	return d.maxWorkers
}
