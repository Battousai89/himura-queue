package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"himura-queue/internal/hashing"
	"himura-queue/pkg/models"
)

type Manager struct {
	shards      map[string]*Shard
	hasher      *hashing.ConsistentHash
	seqCounter  uint64
	lastAck     uint64
	mu          sync.RWMutex
}

func NewManager(shardCount int) *Manager {
	m := &Manager{
		shards: make(map[string]*Shard),
		hasher: hashing.NewConsistentHash(100),
	}
	for i := 0; i < shardCount; i++ {
		shardID := "shard-" + string(rune('0'+i))
		m.shards[shardID] = NewShard(i)
		m.hasher.AddNode(shardID)
	}
	return m
}

func (m *Manager) getShard(queue string) *Shard {
	shardID := m.hasher.GetNode(queue)
	return m.shards[shardID]
}

func (m *Manager) NextID() uint64 {
	return atomic.AddUint64(&m.seqCounter, 1)
}

func (m *Manager) SetLastAck(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id > m.lastAck {
		m.lastAck = id
	}
}

func (m *Manager) GetLastAck() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastAck
}

func (m *Manager) Push(queue string, payload []byte, priority int, delay time.Duration) uint64 {
	m.mu.RLock()
	shard := m.getShard(queue)
	m.mu.RUnlock()
	
	id := m.NextID()
	msg := &models.Message{
		ID:        id,
		Queue:     queue,
		Payload:   payload,
		Priority:  priority,
		Delay:     delay,
		CreatedAt: time.Now(),
	}
	shard.Push(msg)
	return id
}

func (m *Manager) Pop(queue string) *models.Message {
	m.mu.RLock()
	shard := m.getShard(queue)
	m.mu.RUnlock()
	return shard.Pop()
}

func (m *Manager) Len(queue string) int {
	m.mu.RLock()
	shard := m.getShard(queue)
	m.mu.RUnlock()
	return shard.Len()
}

func (m *Manager) GetAllMessages(lastAck uint64) []*models.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var messages []*models.Message
	for _, shard := range m.shards {
		shard.mu.RLock()
		for _, item := range shard.pq.items {
			if item.msg.ID > lastAck {
				messages = append(messages, item.msg)
			}
		}
		for _, item := range shard.dq.items {
			if item.msg.ID > lastAck {
				messages = append(messages, item.msg)
			}
		}
		shard.mu.RUnlock()
	}
	return messages
}

func (m *Manager) MoveDelayed() {
	m.mu.RLock()
	for _, shard := range m.shards {
		shard.MoveDelayed()
	}
	m.mu.RUnlock()
}
