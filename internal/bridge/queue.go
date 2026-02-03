package bridge

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

const MaxQueueSize = 100

var ErrQueueFull = errors.New("injection queue is full")

type Injection struct {
	ID       string
	Text     string
	Priority bool
	SyncChan chan struct{}
}

type Queue struct {
	mu    sync.Mutex
	items []*Injection
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]*Injection, 0),
	}
}

func (q *Queue) Enqueue(text string, priority bool, sync bool) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= MaxQueueSize {
		return "", ErrQueueFull
	}

	inj := &Injection{
		ID:       uuid.New().String(),
		Text:     text,
		Priority: priority,
	}

	if sync {
		inj.SyncChan = make(chan struct{})
	}

	if priority {
		q.items = append([]*Injection{inj}, q.items...)
	} else {
		q.items = append(q.items, inj)
	}

	return inj.ID, nil
}

func (q *Queue) EnqueueWithChan(text string, priority bool, sync bool) (*Injection, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= MaxQueueSize {
		return nil, ErrQueueFull
	}

	inj := &Injection{
		ID:       uuid.New().String(),
		Text:     text,
		Priority: priority,
	}

	if sync {
		inj.SyncChan = make(chan struct{})
	}

	if priority {
		q.items = append([]*Injection{inj}, q.items...)
	} else {
		q.items = append(q.items, inj)
	}

	return inj, nil
}

func (q *Queue) Dequeue() *Injection {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	inj := q.items[0]
	q.items = q.items[1:]
	return inj
}

func (q *Queue) Clear() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := len(q.items)
	q.items = make([]*Injection, 0)
	return count
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

func (q *Queue) Items() []*Injection {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]*Injection, len(q.items))
	copy(result, q.items)
	return result
}
