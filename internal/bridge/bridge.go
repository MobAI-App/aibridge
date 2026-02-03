package bridge

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Bridge struct {
	pty          *PTY
	queue        *Queue
	busyDetector *BusyDetector
	startTime    time.Time
	toolName     string
	verbose      bool
	paranoid     bool
	mu           sync.RWMutex
	running      bool
	injectCh     chan struct{}
	stopCh       chan struct{}
}

func New(command string, args []string, busyPattern string, verbose, paranoid bool) (*Bridge, error) {
	b := &Bridge{
		pty:       NewPTY(command, args),
		queue:     NewQueue(),
		startTime: time.Now(),
		toolName:  command,
		verbose:   verbose,
		paranoid:  paranoid,
		injectCh:  make(chan struct{}, 1),
		stopCh:    make(chan struct{}),
	}

	detector, err := NewBusyDetector(busyPattern, b.triggerInject, verbose)
	if err != nil {
		return nil, fmt.Errorf("invalid busy pattern: %w", err)
	}
	b.busyDetector = detector

	return b, nil
}

func (b *Bridge) Start() error {
	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	err := b.pty.Start(func(line string) {
		b.busyDetector.ProcessLine(line)
	})
	if err != nil {
		return err
	}

	go b.injectionLoop()

	return nil
}

func (b *Bridge) triggerInject() {
	select {
	case b.injectCh <- struct{}{}:
	default:
	}
}

func (b *Bridge) injectionLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case <-b.injectCh:
			b.processQueue()
		}
	}
}

func (b *Bridge) processQueue() {
	if !b.busyDetector.IsIdle() {
		return
	}

	inj := b.queue.Dequeue()
	if inj == nil {
		return
	}

	if b.verbose {
		log.Printf("Injecting text (id=%s): %s", inj.ID, inj.Text)
	}

	b.busyDetector.SetBusy()
	err := b.pty.InjectText(inj.Text, !b.paranoid)
	if err != nil && b.verbose {
		log.Printf("Injection failed: %v", err)
	}

	if inj.SyncChan != nil {
		close(inj.SyncChan)
	}
}

func (b *Bridge) NotifyEnqueue() {
	if b.busyDetector.IsIdle() {
		b.triggerInject()
	}
}

func (b *Bridge) Wait() error {
	err := b.pty.Wait()
	b.mu.Lock()
	b.running = false
	b.mu.Unlock()
	close(b.stopCh)
	return err
}

func (b *Bridge) Close() error {
	return b.pty.Close()
}

func (b *Bridge) Queue() *Queue {
	return b.queue
}

func (b *Bridge) IsIdle() bool {
	return b.busyDetector.IsIdle()
}

func (b *Bridge) IsChildRunning() bool {
	return b.pty.Running()
}

func (b *Bridge) ToolName() string {
	return b.toolName
}

func (b *Bridge) UptimeSeconds() float64 {
	return time.Since(b.startTime).Seconds()
}
