package gworker

import (
	"sync"
)

const (
	defaultWorkerCount = 3
	maxJobCount        = 100000
)

type (
	// Dispatcher managements worker
	Dispatcher struct {
		wg          *sync.WaitGroup
		mu          *sync.Mutex
		workers     []*worker
		workerCount int
		jobs        chan func() error
		joberr      chan error
		finish      chan struct{}
	}

	worker struct {
		dispatcher *Dispatcher
	}
)

// NewDispatcher returns Dispatcher object
func NewDispatcher(workerCount int) *Dispatcher {
	if workerCount < 1 {
		workerCount = defaultWorkerCount
	}

	d := &Dispatcher{
		wg:          new(sync.WaitGroup),
		mu:          new(sync.Mutex),
		workers:     make([]*worker, workerCount),
		workerCount: workerCount,
		jobs:        make(chan func() error, maxJobCount),
		joberr:      make(chan error, maxJobCount),
		finish:      make(chan struct{}, 1),
	}

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
		}

		d.workers[i] = worker
	}

	d.start()

	return d
}

// Start starts workers
func (d *Dispatcher) start() {
	for _, worker := range d.workers {
		go worker.start()
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

	d.mu.Lock()
	defer d.mu.Unlock()

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
		}

		d.workerCount++
		go worker.start()
	}
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
	for job := range w.dispatcher.jobs {
		go func(err error) {
			w.dispatcher.joberr <- err
			w.dispatcher.wg.Done()
		}(job())
	}
}
