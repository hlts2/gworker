package gworker

import (
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
			t.Errorf("tests[%d] - NewDispatcher is nil", i)
		}

		got := d.workerCount

		if test.expected != got {
			t.Errorf("tests[%d] - worker count is wrong: expected: %v, got: %v", i, test.expected, got)
		}
	}
}
