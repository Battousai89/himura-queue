package queue

import (
	"testing"
	"time"

	"himura-queue/pkg/models"
)

func BenchmarkPriorityQueuePush(b *testing.B) {
	pq := NewPriorityQueue()
	msg := &models.Message{ID: 1, Priority: 10, Payload: []byte("test")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.PushMessage(msg)
	}
}

func BenchmarkPriorityQueuePop(b *testing.B) {
	pq := NewPriorityQueue()
	for i := 0; i < b.N; i++ {
		pq.PushMessage(&models.Message{ID: uint64(i), Priority: 10, Payload: []byte("test")})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.PopMessage()
	}
}

func BenchmarkManagerPush(b *testing.B) {
	mgr := NewManager(8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.Push("queue1", []byte("test"), 10, 0)
	}
}

func BenchmarkManagerPop(b *testing.B) {
	mgr := NewManager(8)
	for i := 0; i < b.N; i++ {
		mgr.Push("queue1", []byte("test"), 10, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.Pop("queue1")
	}
}

func BenchmarkDelayedQueue(b *testing.B) {
	dq := NewDelayedQueue()
	msg := &models.Message{ID: 1, Payload: []byte("test")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dq.PushMessage(msg, time.Millisecond)
	}
}
