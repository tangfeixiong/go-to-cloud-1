package dispatcher

import (
	"sync"
)

// A buffered channel that we can send work requests on.
type Scheduler interface {
	Start()
	WithPaylodHandler(handler HandleFunc)
	//Pop() <-chan Job
}

type QueueScheduler struct {
	Dispatcher *Dispatcher
	Queue      chan Job
	wg         *sync.WaitGroup
}

func NewQueueScheduler(maxWorkers int, wg *sync.WaitGroup) Scheduler {
	sch := &QueueScheduler{
		Dispatcher: NewDispatcher(maxWorkers),
		Queue:      make(chan Job),
		wg:         wg,
	}
	sch.create()
	return sch
}

func (qs *QueueScheduler) WithPaylodHandler(handler HandleFunc) {
	qs.Queue <- Job{Payload: Payload{Handler: handler}}
}

func (qs *QueueScheduler) Start() {
	d := qs.Dispatcher
	if qs.wg != nil {
		qs.wg.Add(1)
	}
	go func(sch Scheduler) {
		if qs.wg != nil {
			defer qs.wg.Done()
		}
		for {
			select {
			case job := <-qs.Queue:
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
	}(qs)
}

func (qs *QueueScheduler) create() {
	d := qs.Dispatcher
	// starting n number of workers
	for i := 0; i < d.MaxWorkers(); i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}
}
