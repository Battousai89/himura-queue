package queue

import (
	"testing"
	"time"

	"himura-queue/pkg/models"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	msg1 := &models.Message{ID: 1, Priority: 5, Payload: []byte("low")}
	msg2 := &models.Message{ID: 2, Priority: 10, Payload: []byte("high")}
	msg3 := &models.Message{ID: 3, Priority: 7, Payload: []byte("medium")}

	pq.PushMessage(msg1)
	pq.PushMessage(msg2)
	pq.PushMessage(msg3)

	if pq.LenPublic() != 3 {
		t.Errorf("Expected length 3, got %d", pq.LenPublic())
	}

	got := pq.PopMessage()
	if got.ID != 2 {
		t.Errorf("Expected highest priority message ID 2, got %d", got.ID)
	}

	got = pq.PopMessage()
	if got.ID != 3 {
		t.Errorf("Expected medium priority message ID 3, got %d", got.ID)
	}

	got = pq.PopMessage()
	if got.ID != 1 {
		t.Errorf("Expected lowest priority message ID 1, got %d", got.ID)
	}
}

func TestDelayedQueue(t *testing.T) {
	dq := NewDelayedQueue()

	msg := &models.Message{ID: 1, Payload: []byte("delayed")}
	dq.PushMessage(msg, 100*time.Millisecond)

	if dq.LenPublic() != 1 {
		t.Errorf("Expected length 1, got %d", dq.LenPublic())
	}

	if dq.PopReady() != nil {
		t.Error("Expected nil for not-yet-ready message")
	}

	time.Sleep(150 * time.Millisecond)

	got := dq.PopReady()
	if got == nil {
		t.Error("Expected ready message, got nil")
	}
	if got.ID != 1 {
		t.Errorf("Expected message ID 1, got %d", got.ID)
	}
}

func TestShard(t *testing.T) {
	shard := NewShard(0)

	msg1 := &models.Message{ID: 1, Priority: 5, Payload: []byte("immediate")}
	msg2 := &models.Message{ID: 2, Priority: 10, Payload: []byte("delayed"), Delay: 50 * time.Millisecond}

	shard.Push(msg1)
	shard.Push(msg2)

	got := shard.Pop()
	if got.ID != 1 {
		t.Errorf("Expected immediate message ID 1, got %d", got.ID)
	}

	time.Sleep(100 * time.Millisecond)
	shard.MoveDelayed()

	got = shard.Pop()
	if got == nil || got.ID != 2 {
		t.Errorf("Expected delayed message ID 2, got %v", got)
	}
}

func TestManager(t *testing.T) {
	mgr := NewManager(4)

	id := mgr.Push("queue1", []byte("test"), 10, 0)
	if id == 0 {
		t.Error("Expected non-zero message ID")
	}

	msg := mgr.Pop("queue1")
	if msg == nil {
		t.Error("Expected message, got nil")
	}
	if msg.ID != id {
		t.Errorf("Expected message ID %d, got %d", id, msg.ID)
	}
}
