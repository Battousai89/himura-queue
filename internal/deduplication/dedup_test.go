package deduplication

import (
	"os"
	"testing"
	"time"
)

func TestDeduplicator(t *testing.T) {
	dedup := &Deduplicator{
		seen:    make(map[uint64]Entry),
		ttl:     100 * time.Millisecond,
		path:    "",
		lastAck: 0,
	}

	if dedup.IsDuplicate(1) {
		t.Error("First occurrence should not be duplicate")
	}

	if !dedup.IsDuplicate(1) {
		t.Error("Second occurrence should be duplicate")
	}

	if dedup.IsDuplicate(2) {
		t.Error("Different ID should not be duplicate")
	}
}

func TestDeduplicatorTTL(t *testing.T) {
	dedup := &Deduplicator{
		seen:    make(map[uint64]Entry),
		ttl:     50 * time.Millisecond,
		path:    "",
		lastAck: 0,
	}

	dedup.IsDuplicate(1)
	time.Sleep(100 * time.Millisecond)
	dedup.cleanup()

	if dedup.IsDuplicate(1) {
		t.Error("After TTL should not be duplicate")
	}
}

func TestDeduplicatorSaveLoad(t *testing.T) {
	tmpFile := "test_dedup_save.bin"
	defer os.Remove(tmpFile)
	
	dedup := &Deduplicator{
		seen:    make(map[uint64]Entry),
		ttl:     5 * time.Minute,
		path:    tmpFile,
		lastAck: 0,
	}
	dedup.IsDuplicate(1)
	dedup.IsDuplicate(2)
	dedup.IsDuplicate(3)
	dedup.save()
	
	dedup2 := &Deduplicator{
		seen:    make(map[uint64]Entry),
		ttl:     5 * time.Minute,
		path:    tmpFile,
		lastAck: 0,
	}
	dedup2.load()
	
	if !dedup2.IsDuplicate(1) {
		t.Error("Expected 1 to be duplicate after load")
	}
	if !dedup2.IsDuplicate(2) {
		t.Error("Expected 2 to be duplicate after load")
	}
	if !dedup2.IsDuplicate(3) {
		t.Error("Expected 3 to be duplicate after load")
	}
	if dedup2.IsDuplicate(4) {
		t.Error("Expected 4 to not be duplicate")
	}
}
