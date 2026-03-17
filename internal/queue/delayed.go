package queue

import (
	"container/heap"
	"sync"
	"time"

	"himura-queue/pkg/models"
)

type DelayedItem struct {
	msg   *models.Message
	index int
	time  time.Time
}

type DelayedQueue struct {
	items []*DelayedItem
	mu    sync.RWMutex
}

func NewDelayedQueue() *DelayedQueue {
	return &DelayedQueue{
		items: make([]*DelayedItem, 0),
	}
}

func (dq *DelayedQueue) Len() int {
	return len(dq.items)
}

func (dq *DelayedQueue) Less(i, j int) bool {
	return dq.items[i].time.Before(dq.items[j].time)
}

func (dq *DelayedQueue) Swap(i, j int) {
	dq.items[i], dq.items[j] = dq.items[j], dq.items[i]
	dq.items[i].index = i
	dq.items[j].index = j
}

func (dq *DelayedQueue) Push(x interface{}) {
	item := x.(*DelayedItem)
	item.index = len(dq.items)
	dq.items = append(dq.items, item)
}

func (dq *DelayedQueue) Pop() interface{} {
	old := dq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	dq.items = old[0 : n-1]
	return item
}

func (dq *DelayedQueue) PushMessage(msg *models.Message, delay time.Duration) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	item := &DelayedItem{
		msg:  msg,
		time: time.Now().Add(delay),
	}
	heap.Push(dq, item)
}

func (dq *DelayedQueue) PopReady() *models.Message {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	if len(dq.items) == 0 {
		return nil
	}
	if dq.items[0].time.After(time.Now()) {
		return nil
	}
	item := heap.Pop(dq).(*DelayedItem)
	return item.msg
}

func (dq *DelayedQueue) NextReadyTime() time.Time {
	dq.mu.RLock()
	defer dq.mu.RUnlock()
	if len(dq.items) == 0 {
		return time.Time{}
	}
	return dq.items[0].time
}

func (dq *DelayedQueue) LenPublic() int {
	dq.mu.RLock()
	defer dq.mu.RUnlock()
	return len(dq.items)
}
