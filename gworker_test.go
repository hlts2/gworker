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

		got := d.workerCount

		if test.expected != got {
			t.Errorf("tests[%d] - worker count is wrong. expected: %v, got: %v", i, test.expected, got)
		}
	}
}

func TestStart(t *testing.T) {
	workerCount := 10

	d := NewDispatcher(workerCount)
	d.Start()

	got := runtime.NumGoroutine() - 2
	if workerCount != got {
		t.Errorf("Starting worker count is wrong. expected: %v, got: %v", workerCount, got)
	}
}
