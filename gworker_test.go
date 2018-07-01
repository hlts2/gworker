package gworker

import (
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

func TestStart(t *testing.T) {
	test := struct {
		workerCount int
		expected    int
	}{
		workerCount: 10,
		expected:    10,
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	got := runtime.NumGoroutine() - 2
	if test.expected != got {
		t.Errorf("Starting worker count is wrong. expected: %v, got: %v", test.expected, got)
	}

	d.Stop()
}

func TestStop(t *testing.T) {
	test := struct {
		workerCount int
		expected    int
	}{
		workerCount: 5,
		expected:    0,
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	d.Stop()

	got := runtime.NumGoroutine() - 2
	if test.expected != got {
		t.Errorf("Stop worker count is wrong. expected: %v, got: %v", test.expected, got)
	}
}

func TestUpScale(t *testing.T) {
	test := struct {
		workerCount  int
		upScaleCount int
		expected     int
	}{
		workerCount:  10,
		upScaleCount: 5,
		expected:     15,
	}

	d := NewDispatcher(test.workerCount)
	d.Start()

	d.UpScale(test.upScaleCount)

	got := runtime.NumGoroutine() - 2
	if test.expected != got {
		t.Errorf("Upscale is wrong. expected: %v, got: %v", test.upScaleCount, got)
	}

	d.Stop()
}
