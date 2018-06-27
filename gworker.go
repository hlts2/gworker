package gworker

import (
	"sync"
	"sync/atomic"
)

const (
	defaultWorkerCount = 3
	maxJobCount        = 100000
)

type (
	// Dispatcher managements worker
	Dispatcher struct {
		wg          *sync.WaitGroup
		has         int32
		workers     []*worker
		workerCount int
		runnig      bool
		jobs        chan func() error
		joberr      chan error
		finish      chan struct{}
	}

	worker struct {
		dispatcher *Dispatcher
		runnig     bool
	}
)

// NewDispatcher returns Dispatcher object
func NewDispatcher(workerCount int) *Dispatcher {
	if workerCount < 1 {
		workerCount = defaultWorkerCount
	}

	d := &Dispatcher{
		wg:          new(sync.WaitGroup),
		has:         0,
		workers:     make([]*worker, workerCount),
		workerCount: workerCount,
		runnig:      false,
		jobs:        make(chan func() error, maxJobCount),
		joberr:      make(chan error),
		finish:      make(chan struct{}, 1),
	}

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
			runnig:     false,
		}

		d.workers[i] = worker
	}

	return d
}

// Start starts workers
func (d *Dispatcher) Start() {
	if d.runnig {
		return
	}

	for _, worker := range d.workers {
		if !worker.runnig {
			go worker.start()
		}
	}
}

// Add adds job
func (d *Dispatcher) Add(job func() error) {
	d.wg.Add(1)
	d.jobs <- job
}

// StartJobObserver monitors jobs. Then, when all the jobs are completed, the notification is transmitted
func (d *Dispatcher) StartJobObserver() {
	go func() {
		for {
			if len(d.jobs) > 0 {
				d.wg.Wait()
				d.finish <- struct{}{}
				return
			}
		}
	}()
}

// UpScale scales up the numer of worker
func (d *Dispatcher) UpScale(workerCount int) {
	if workerCount < 1 {
		workerCount = defaultWorkerCount
	}

	for {
		if d.has == 0 && atomic.CompareAndSwapInt32(&d.has, 0, 1) {
			break
		}
	}

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
			runnig:     false,
		}

		d.workerCount++
		go worker.start()
	}

	d.has = 0
}

// JobError returns channel for job error
func (d *Dispatcher) JobError() <-chan error {
	return d.joberr
}

// Finish returns channel when all jobs are finished
func (d *Dispatcher) Finish() <-chan struct{} {
	return d.finish
}

func (w *worker) start() {
	w.runnig = true
	for job := range w.dispatcher.jobs {
		go func(err error) {
			w.dispatcher.joberr <- err
			w.dispatcher.wg.Done()
		}(job())
	}
	w.runnig = false
}
