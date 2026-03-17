package queue

import (
	"container/heap"
	"sync"

	"himura-queue/pkg/models"
)

type PriorityItem struct {
	msg      *models.Message
	index    int
	priority int
}

type PriorityQueue struct {
	items []*PriorityItem
	mu    sync.RWMutex
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*PriorityItem, 0),
	}
}

func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.items[i].priority > pq.items[j].priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*PriorityItem)
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	pq.items = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) PushMessage(msg *models.Message) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	item := &PriorityItem{
		msg:      msg,
		priority: msg.Priority,
	}
	heap.Push(pq, item)
}

func (pq *PriorityQueue) PopMessage() *models.Message {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.items) == 0 {
		return nil
	}
	item := heap.Pop(pq).(*PriorityItem)
	return item.msg
}

func (pq *PriorityQueue) Peek() *models.Message {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0].msg
}

func (pq *PriorityQueue) LenPublic() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}
