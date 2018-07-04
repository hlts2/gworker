package gworker

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultWorkerCount = 3
	maxJobCount        = 100000
)

type (
	// Dispatcher managements worker
	Dispatcher struct {
		wg          *sync.WaitGroup
		workers     []*worker
		workerCount int
		sflg        int32
		runnig      bool
		scaling     bool
		observing   bool
		stopWorkers chan bool
		jobs        chan func() error
		joberr      chan error
		finish      chan struct{}
	}

	worker struct {
		id         int
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
		workers:     make([]*worker, workerCount),
		workerCount: workerCount,
		sflg:        0,
		runnig:      false,
		scaling:     false,
		observing:   false,
		stopWorkers: make(chan bool),
		jobs:        make(chan func() error, maxJobCount),
		joberr:      make(chan error),
		finish:      make(chan struct{}, 1),
	}

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			id:         i,
			dispatcher: d,
			runnig:     false,
		}

		d.workers[i] = worker
	}

	return d
}

// Start starts workers
func (d *Dispatcher) Start() *Dispatcher {
	if d.runnig {
		return d
	}

	d.runnig = true

	for _, worker := range d.workers {
		worker.start()
	}
	return d
}

// Stop stops workers.
func (d *Dispatcher) Stop() *Dispatcher {
	if !d.runnig {
		return d
	}

	for i := 0; i < d.workerCount; i++ {
		// send stop event of worker
		d.stopWorkers <- true
	}

	// delay until goroutine is collected in GC
	time.Sleep(1 * time.Millisecond)

	d.runnig = false

	return d
}

// Add adds job
func (d *Dispatcher) Add(job func() error) {
	d.wg.Add(1)
	d.jobs <- job
}

// StartJobObserver monitors jobs. Then, when all the jobs are completed, the notification is transmitted
func (d *Dispatcher) StartJobObserver() {
	if d.observing {
		return
	}

	d.observing = true

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

// GetWorkerCount returns the number of workers
func (d *Dispatcher) GetWorkerCount() int {
	for {
		if !d.scaling {
			return len(d.workers)
		}
	}
}

// UpScale scales up the numer of worker
func (d *Dispatcher) UpScale(workerCount int) *Dispatcher {
	if workerCount < 1 {
		workerCount = defaultWorkerCount
	}

	for {
		if d.sflg == 0 && atomic.CompareAndSwapInt32(&d.sflg, 0, 1) {
			break
		}
	}

	d.scaling = true

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
			runnig:     false,
		}

		d.workerCount++
		go worker.start()
	}

	d.sflg = 0
	d.scaling = false

	return d
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
	go func() {
		for {
			select {
			case job, _ := <-w.dispatcher.jobs:
				w.run(job)
			case _ = <-w.dispatcher.stopWorkers:
				return
			}
		}
	}()
}

func (w *worker) run(job func() error) {
	go func(err error) {
		w.dispatcher.joberr <- err
		w.dispatcher.wg.Done()
	}(job())
}
