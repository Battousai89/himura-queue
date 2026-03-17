package deduplication

import (
	"encoding/binary"
	"os"
	"sync"
	"time"
)

type Entry struct {
	expiresAt time.Time
}

type Deduplicator struct {
	seen     map[uint64]Entry
	mu       sync.RWMutex
	ttl      time.Duration
	path     string
	lastAck  uint64
}

func NewDeduplicator(ttl time.Duration, path string) *Deduplicator {
	d := &Deduplicator{
		seen:    make(map[uint64]Entry),
		ttl:     ttl,
		path:    path,
		lastAck: 0,
	}
	d.load()
	go d.cleanupLoop()
	return d
}

func (d *Deduplicator) IsDuplicate(id uint64) bool {
	d.mu.RLock()
	if _, exists := d.seen[id]; exists {
		d.mu.RUnlock()
		return true
	}
	d.mu.RUnlock()

	d.mu.Lock()
	d.seen[id] = Entry{
		expiresAt: time.Now().Add(d.ttl),
	}
	if id > d.lastAck {
		d.lastAck = id
	}
	d.mu.Unlock()
	d.save()
	return false
}

func (d *Deduplicator) GetLastAck() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastAck
}

func (d *Deduplicator) save() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.saveLocked()
}

func (d *Deduplicator) saveLocked() {
	if d.path == "" {
		return
	}
	f, err := os.Create(d.path)
	if err != nil {
		return
	}
	defer f.Close()

	binary.Write(f, binary.BigEndian, d.lastAck)
	count := uint32(len(d.seen))
	binary.Write(f, binary.BigEndian, count)
	for id, entry := range d.seen {
		binary.Write(f, binary.BigEndian, id)
		binary.Write(f, binary.BigEndian, entry.expiresAt.UnixNano())
	}
}

func (d *Deduplicator) load() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	f, err := os.Open(d.path)
	if err != nil {
		return
	}
	defer f.Close()
	
	binary.Read(f, binary.BigEndian, &d.lastAck)
	var count uint32
	binary.Read(f, binary.BigEndian, &count)
	for i := uint32(0); i < count; i++ {
		var id uint64
		var expires int64
		binary.Read(f, binary.BigEndian, &id)
		binary.Read(f, binary.BigEndian, &expires)
		if time.Unix(0, expires).After(time.Now()) {
			d.seen[id] = Entry{
				expiresAt: time.Unix(0, expires),
			}
		}
	}
}

func (d *Deduplicator) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		d.cleanup()
	}
}

func (d *Deduplicator) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for id, entry := range d.seen {
		if entry.expiresAt.Before(now) {
			delete(d.seen, id)
		}
	}
	d.saveLocked()
}
