package worker

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolSubmit(t *testing.T) {
	done := make(chan struct{})
	var counter int32

	pool := NewPool(10, 2, 10, 5*time.Second)

	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	go func() {
		for atomic.LoadInt32(&counter) < 10 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for tasks")
	}

	pool.Shutdown()

	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("Expected 10 tasks executed, got %d", atomic.LoadInt32(&counter))
	}
}

func TestPoolScaleUp(t *testing.T) {
	done := make(chan struct{})
	var counter int32

	pool := NewPool(100, 2, 10, 5*time.Second)

	for i := 0; i < 20; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	go func() {
		for atomic.LoadInt32(&counter) < 20 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for tasks")
	}

	pool.Shutdown()

	if atomic.LoadInt32(&counter) != 20 {
		t.Errorf("Expected 20 tasks executed, got %d", atomic.LoadInt32(&counter))
	}
}
