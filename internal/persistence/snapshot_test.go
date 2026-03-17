package persistence

import (
	"os"
	"testing"
	"time"

	"himura-queue/pkg/models"
)

func TestSnapshotSaveLoad(t *testing.T) {
	tmpFile := "test_snapshot.bin"
	defer os.Remove(tmpFile)

	snapshotter := NewSnapshotter(tmpFile, time.Minute)

	messages := []*models.Message{
		{
			ID:        1,
			Queue:     "queue1",
			Payload:   []byte("hello"),
			Priority:  10,
			Delay:     0,
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Queue:     "queue2",
			Payload:   []byte("world"),
			Priority:  5,
			Delay:     time.Second,
			CreatedAt: time.Now(),
		},
	}

	if err := snapshotter.Save(messages); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := snapshotter.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if len(loaded) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(loaded))
	}

	for i, msg := range loaded {
		if msg.ID != messages[i].ID {
			t.Errorf("Message %d ID mismatch: got %d, want %d", i, msg.ID, messages[i].ID)
		}
		if msg.Queue != messages[i].Queue {
			t.Errorf("Message %d Queue mismatch: got %s, want %s", i, msg.Queue, messages[i].Queue)
		}
		if string(msg.Payload) != string(messages[i].Payload) {
			t.Errorf("Message %d Payload mismatch: got %s, want %s", i, msg.Payload, messages[i].Payload)
		}
	}
}

func TestSnapshotLoadNonExistent(t *testing.T) {
	snapshotter := NewSnapshotter("nonexistent.bin", time.Minute)

	messages, err := snapshotter.Load()
	if err != nil {
		t.Fatalf("Load should return nil for non-existent file: %v", err)
	}
	if messages != nil {
		t.Error("Expected nil messages for non-existent file")
	}
}
