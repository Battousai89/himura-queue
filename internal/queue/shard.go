package queue

import (
	"sync"

	"himura-queue/pkg/models"
)

type Shard struct {
	id       int
	pq       *PriorityQueue
	dq       *DelayedQueue
	mu       sync.RWMutex
}

func NewShard(id int) *Shard {
	return &Shard{
		id: id,
		pq: NewPriorityQueue(),
		dq: NewDelayedQueue(),
	}
}

func (s *Shard) Push(msg *models.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg.Delay > 0 {
		s.dq.PushMessage(msg, msg.Delay)
	} else {
		s.pq.PushMessage(msg)
	}
}

func (s *Shard) Pop() *models.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg := s.dq.PopReady(); msg != nil {
		return msg
	}
	return s.pq.PopMessage()
}

func (s *Shard) Peek() *models.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if msg := s.pq.Peek(); msg != nil {
		return msg
	}
	return nil
}

func (s *Shard) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pq.LenPublic() + s.dq.LenPublic()
}

func (s *Shard) MoveDelayed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		msg := s.dq.PopReady()
		if msg == nil {
			break
		}
		s.pq.PushMessage(msg)
	}
}
