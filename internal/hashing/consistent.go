package hashing

import (
	"hash/fnv"
	"sort"
	"sync"
)

type ConsistentHash struct {
	hashes      []uint32
	keys        map[uint32]string
	virtualNodes int
	mu          sync.RWMutex
}

func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		hashes:       make([]uint32, 0),
		keys:         make(map[uint32]string),
		virtualNodes: virtualNodes,
	}
}

func (ch *ConsistentHash) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (ch *ConsistentHash) AddNode(node string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for i := 0; i < ch.virtualNodes; i++ {
		key := ch.hash(node + "#" + string(rune(i)))
		ch.hashes = append(ch.hashes, key)
		ch.keys[key] = node
	}
	sort.Slice(ch.hashes, func(i, j int) bool {
		return ch.hashes[i] < ch.hashes[j]
	})
}

func (ch *ConsistentHash) RemoveNode(node string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for i := 0; i < ch.virtualNodes; i++ {
		key := ch.hash(node + "#" + string(rune(i)))
		idx := sort.Search(len(ch.hashes), func(i int) bool {
			return ch.hashes[i] >= key
		})
		if idx < len(ch.hashes) && ch.hashes[idx] == key {
			ch.hashes = append(ch.hashes[:idx], ch.hashes[idx+1:]...)
			delete(ch.keys, key)
		}
	}
}

func (ch *ConsistentHash) GetNode(key string) string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	if len(ch.hashes) == 0 {
		return ""
	}
	hash := ch.hash(key)
	idx := sort.Search(len(ch.hashes), func(i int) bool {
		return ch.hashes[i] >= hash
	})
	if idx == len(ch.hashes) {
		idx = 0
	}
	return ch.keys[ch.hashes[idx]]
}
