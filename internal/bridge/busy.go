package bridge

import (
	"log"
	"sync"
	"time"
)

type BusyDetector struct {
	mu          sync.RWMutex
	idle        bool
	lastOutput  time.Time
	onIdle      func()
	verbose     bool
	idleTimeout time.Duration
}

func NewBusyDetector(pattern string, onIdle func(), verbose bool) (*BusyDetector, error) {
	d := &BusyDetector{
		idle:        true,
		onIdle:      onIdle,
		verbose:     verbose,
		idleTimeout: 500 * time.Millisecond,
	}

	go d.checkIdleLoop()

	return d, nil
}

func (d *BusyDetector) checkIdleLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		wasIdle := d.idle
		if !d.idle && time.Since(d.lastOutput) > d.idleTimeout {
			d.idle = true
			if d.verbose {
				log.Printf("Idle timeout - no output for %v", d.idleTimeout)
			}
		}
		isIdle := d.idle
		d.mu.Unlock()

		if isIdle && !wasIdle && d.onIdle != nil {
			d.onIdle()
		}
	}
}

func (d *BusyDetector) ProcessLine(line string) {
	if d.verbose {
		log.Printf("PTY line: %q", line)
	}
	d.mu.Lock()
	d.idle = false
	d.lastOutput = time.Now()
	d.mu.Unlock()
}

func (d *BusyDetector) IsIdle() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.idle
}

func (d *BusyDetector) SetBusy() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.idle = false
	d.lastOutput = time.Now()
}
