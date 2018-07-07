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
		wg        *sync.WaitGroup
		workers   []*worker
		sflg      int32
		runnig    bool
		scaling   bool
		observing bool
		start     chan bool
		jobs      chan func() error
		joberr    chan error
		finish    chan struct{}
	}

	worker struct {
		dispatcher *Dispatcher
		runnig     bool
		stop       chan bool
	}
)

// NewDispatcher returns Dispatcher object
func NewDispatcher(workerCount int) *Dispatcher {
	if workerCount < 1 {
		workerCount = defaultWorkerCount
	}

	d := &Dispatcher{
		wg:        new(sync.WaitGroup),
		workers:   make([]*worker, workerCount),
		sflg:      0,
		runnig:    false,
		scaling:   false,
		observing: false,
		start:     make(chan bool, 1),
		jobs:      make(chan func() error, maxJobCount),
		joberr:    make(chan error),
		finish:    make(chan struct{}, 1),
	}

	for i := 0; i < workerCount; i++ {
		worker := &worker{
			dispatcher: d,
			runnig:     false,
			stop:       make(chan bool),
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

	d.start <- true
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

	for _, worker := range d.workers {
		// sends stop event of worker
		worker.stop <- true
		worker.runnig = false
	}

	// delays until goroutine is collected in GC
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
		<-d.start
		d.wg.Wait()
		d.finish <- struct{}{}
		return
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
	if !d.runnig {
		return d
	}

	if workerCount < 1 {
		return d
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
			stop:       make(chan bool),
		}

		worker.start()
		d.workers = append(d.workers, worker)
	}

	d.sflg = 0
	d.scaling = false

	return d
}

// DownScale scales down the numer of worker
func (d *Dispatcher) DownScale(workerCount int) *Dispatcher {
	if !d.runnig {
		return d
	}

	if workerCount < 1 || workerCount > len(d.workers) {
		return d
	}

	for {
		if d.sflg == 0 && atomic.CompareAndSwapInt32(&d.sflg, 0, 1) {
			break
		}
	}

	d.scaling = true

	for i := len(d.workers) - 1; i >= len(d.workers)-workerCount; i-- {
		// send stop event of worker
		d.workers[i].stop <- true
		d.workers[i].runnig = false
	}

	// delay until goroutine is collected in GC
	time.Sleep(1 * time.Millisecond)

	d.workers = d.workers[:len(d.workers)-workerCount]

	d.sflg = 0
	d.scaling = false

	return d
}

// JobError returns error of job
func (d *Dispatcher) JobError() <-chan error {
	return d.joberr
}

// Finish notifies when all jobs are finished
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
			case _ = <-w.stop:
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
