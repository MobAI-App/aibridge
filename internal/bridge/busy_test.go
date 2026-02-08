package bridge

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBusyDetectorOutput(t *testing.T) {
	var called int32
	d, err := NewBusyDetector("", func() {
		atomic.AddInt32(&called, 1)
	}, false)
	if err != nil {
		t.Fatalf("NewBusyDetector failed: %v", err)
	}

	if !d.IsIdle() {
		t.Error("Should be idle initially")
	}

	d.ProcessLine("any output at all")
	if d.IsIdle() {
		t.Error("Should be busy after any output")
	}

	time.Sleep(600 * time.Millisecond)
	if !d.IsIdle() {
		t.Error("Should be idle after timeout")
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Error("OnIdle callback should have been called")
	}
}

func TestBusyDetectorSetBusy(t *testing.T) {
	d, _ := NewBusyDetector("", nil, false)

	if !d.IsIdle() {
		t.Error("Should be idle initially")
	}

	d.SetBusy()
	if d.IsIdle() {
		t.Error("Should not be idle after SetBusy")
	}
}
