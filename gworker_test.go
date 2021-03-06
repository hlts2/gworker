package gworker

import (
	"errors"
	"runtime"
	"testing"
)

func TestNewDispatcher(t *testing.T) {
	tests := []struct {
		workerCount int
		expected    int
	}{
		{
			workerCount: 1,
			expected:    1,
		},
		{
			workerCount: 100,
			expected:    100,
		},
		{
			workerCount: 0,
			expected:    3,
		},
		{
			workerCount: -1,
			expected:    3,
		},
	}

	for i, test := range tests {
		d := NewDispatcher(test.workerCount)
		if d == nil {
			t.Errorf("tests[%d] - NewDispatcher is nil.", i)
		}
	}
}

func TestAdd(t *testing.T) {
	test := struct {
		workerCount int
		jobCount    int
		expected    int
	}{
		workerCount: 5,
		jobCount:    100,
		expected:    100,
	}

	d := NewDispatcher(test.workerCount)

	for i := 0; i < test.jobCount; i++ {
		d.Add(func() error {
			return nil
		})
	}

	got := len(d.jobs)
	if test.expected != got {
		t.Errorf("Add job count is wrong. expected: %v, got: %v", test.expected, got)
	}
}

func TestStartAndStop(t *testing.T) {
	test := struct {
		workerCount      int
		expectedForStart int
		expectedForStop  int
	}{
		workerCount:      10,
		expectedForStart: 10,
		expectedForStop:  0,
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	got := runtime.NumGoroutine() - 2
	if test.expectedForStart != got {
		t.Errorf("Starting worker count is wrong. expected: %v, got: %v", test.expectedForStart, got)
	}

	d.Stop()

	got = runtime.NumGoroutine() - 2
	if test.expectedForStop != got {
		t.Errorf("Stop worker count is wrong. expected: %v, got: %v", test.expectedForStop, got)
	}
}

func TestUpScale(t *testing.T) {
	tests := []struct {
		workerCount  int
		upScaleCount int
		expected     int
	}{
		{
			workerCount:  10,
			upScaleCount: 5,
			expected:     15,
		},
		{
			workerCount:  10,
			upScaleCount: 1,
			expected:     11,
		},
		{
			workerCount:  10,
			upScaleCount: 0,
			expected:     10,
		},
		{
			workerCount:  10,
			upScaleCount: -1,
			expected:     10,
		},
	}

	for i, test := range tests {
		d := NewDispatcher(test.workerCount)
		d.Start()

		d.UpScale(test.upScaleCount)

		got := runtime.NumGoroutine() - 2
		if test.expected != got {
			t.Errorf("tests[%d] - Upscale is wrong. expected: %v, got: %v", i, test.upScaleCount, got)
		}

		d.Stop()
	}
}

func TestDownScale(t *testing.T) {
	tests := []struct {
		workerCount    int
		downScaleCount int
		expected       int
	}{
		{
			workerCount:    10,
			downScaleCount: 5,
			expected:       5,
		},
		{
			workerCount:    10,
			downScaleCount: 1,
			expected:       9,
		},
		{
			workerCount:    10,
			downScaleCount: 0,
			expected:       10,
		},
		{
			workerCount:    10,
			downScaleCount: -1,
			expected:       10,
		},
	}

	for i, test := range tests {
		d := NewDispatcher(test.workerCount)
		d.Start()

		d.DownScale(test.downScaleCount)

		got := runtime.NumGoroutine() - 2
		if test.expected != got {
			t.Errorf("tests[%d] - Downscale is wrong. expected: %v, got: %v", i, test.downScaleCount, got)
		}

		d.Stop()
	}
}

func TestGetWorkerCount(t *testing.T) {
	test := struct {
		workerCount int
		expected    int
	}{
		workerCount: 10,
		expected:    10,
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	got := d.GetWorkerCount()

	if test.expected != got {
		t.Errorf("GetWorkerCount is wrong. expected: %v, got: %v", test.expected, got)
	}

	d.Stop()
}

func TestJobError(t *testing.T) {
	test := struct {
		workerCount int
		expected    error
	}{
		workerCount: 10,
		expected:    errors.New("gworker error"),
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	d.Add(func() error {
		return errors.New("gworker error")
	})

	got := <-d.JobError()

	if test.expected.Error() != got.Error() {
		t.Errorf("JobError is wrong. expected: %v, got: %v", test.expected, got)
	}

	d.Stop()
}

func TestFinish(t *testing.T) {
	test := struct {
		workerCount int
		jobCount    int
		expected    struct{}
	}{
		workerCount: 10,
		jobCount:    10,
		expected:    struct{}{},
	}

	d := NewDispatcher(test.workerCount)
	d.StartJobObserver()

	for i := 0; i < test.jobCount; i++ {
		d.Add(func() error {
			return nil
		})
	}

	d.Start()

END_LOOP:
	for {
		select {
		case _ = <-d.JobError():
		case got := <-d.Finish():
			if test.expected != got {
				t.Errorf("Finish is wrong: expected: %v, got: %v", test.expected, got)
			}
			break END_LOOP
		}
	}

	if len(d.jobs) != 0 {
		t.Errorf("jobs is wrong: expected: %v, got: %v", 0, len(d.jobs))
	}

	d.Stop()
}
