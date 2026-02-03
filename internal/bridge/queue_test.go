package bridge

import "testing"

func TestQueueEnqueueDequeue(t *testing.T) {
	q := NewQueue()

	id, err := q.Enqueue("test1", false, false)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if id == "" {
		t.Error("Enqueue returned empty ID")
	}
	if q.Len() != 1 {
		t.Errorf("Len() = %d, want 1", q.Len())
	}

	inj := q.Dequeue()
	if inj == nil || inj.Text != "test1" {
		t.Errorf("Dequeue() = %v, want Text=%q", inj, "test1")
	}
	if q.Len() != 0 {
		t.Errorf("Len() = %d, want 0", q.Len())
	}
}

func TestQueuePriority(t *testing.T) {
	q := NewQueue()

	if _, err := q.Enqueue("normal1", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("normal2", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("priority", true, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	inj := q.Dequeue()
	if inj.Text != "priority" {
		t.Errorf("Expected priority item first, got %q", inj.Text)
	}

	inj = q.Dequeue()
	if inj.Text != "normal1" {
		t.Errorf("Expected normal1 second, got %q", inj.Text)
	}
}

func TestQueueFIFO(t *testing.T) {
	q := NewQueue()

	if _, err := q.Enqueue("first", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("second", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("third", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	order := []string{"first", "second", "third"}
	for _, expected := range order {
		inj := q.Dequeue()
		if inj.Text != expected {
			t.Errorf("Got %q, want %q", inj.Text, expected)
		}
	}
}

func TestQueueMaxSize(t *testing.T) {
	q := NewQueue()

	for i := 0; i < MaxQueueSize; i++ {
		_, err := q.Enqueue("item", false, false)
		if err != nil {
			t.Fatalf("Enqueue %d failed: %v", i, err)
		}
	}

	_, err := q.Enqueue("overflow", false, false)
	if err != ErrQueueFull {
		t.Errorf("Expected ErrQueueFull, got %v", err)
	}
}

func TestQueueClear(t *testing.T) {
	q := NewQueue()

	if _, err := q.Enqueue("item1", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("item2", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("item3", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	count := q.Clear()
	if count != 3 {
		t.Errorf("Clear() = %d, want 3", count)
	}
	if q.Len() != 0 {
		t.Errorf("Len() after Clear = %d, want 0", q.Len())
	}
}

func TestQueueDequeueEmpty(t *testing.T) {
	q := NewQueue()

	inj := q.Dequeue()
	if inj != nil {
		t.Error("Dequeue on empty queue should return nil")
	}
}

func TestQueueSyncChan(t *testing.T) {
	q := NewQueue()

	inj, err := q.EnqueueWithChan("sync", false, true)
	if err != nil {
		t.Fatalf("EnqueueWithChan failed: %v", err)
	}
	if inj.SyncChan == nil {
		t.Error("SyncChan should not be nil for sync mode")
	}

	inj2, err := q.EnqueueWithChan("async", false, false)
	if err != nil {
		t.Fatalf("EnqueueWithChan failed: %v", err)
	}
	if inj2.SyncChan != nil {
		t.Error("SyncChan should be nil for async mode")
	}
}

func TestQueueItems(t *testing.T) {
	q := NewQueue()

	if _, err := q.Enqueue("item1", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	if _, err := q.Enqueue("item2", false, false); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	items := q.Items()
	if len(items) != 2 {
		t.Errorf("Items() len = %d, want 2", len(items))
	}

	items[0] = nil
	if q.Len() != 2 {
		t.Error("Items() should return a copy")
	}
}
