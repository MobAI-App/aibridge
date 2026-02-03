package bridge

import (
	"log"
	"regexp"
	"sync"
	"time"
)

type BusyDetector struct {
	pattern     *regexp.Regexp
	mu          sync.RWMutex
	idle        bool
	lastBusy    time.Time
	onIdle      func()
	verbose     bool
	idleTimeout time.Duration
}

func NewBusyDetector(pattern string, onIdle func(), verbose bool) (*BusyDetector, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	d := &BusyDetector{
		pattern:     re,
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
		if !d.idle && time.Since(d.lastBusy) > d.idleTimeout {
			d.idle = true
			if d.verbose {
				log.Printf("Idle timeout - no busy signal for %v", d.idleTimeout)
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
	if d.pattern.MatchString(line) {
		d.mu.Lock()
		d.idle = false
		d.lastBusy = time.Now()
		d.mu.Unlock()
		if d.verbose {
			log.Printf("Busy pattern matched: %q", line)
		}
	}
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
	d.lastBusy = time.Now()
}
