package hashing

import (
	"testing"
)

func TestConsistentHash(t *testing.T) {
	ch := NewConsistentHash(100)

	ch.AddNode("node1")
	ch.AddNode("node2")
	ch.AddNode("node3")

	node := ch.GetNode("test-key")
	if node == "" {
		t.Error("Expected non-empty node")
	}

	for i := 0; i < 100; i++ {
		got := ch.GetNode("test-key")
		if got != node {
			t.Errorf("Inconsistent hashing: got %s, want %s", got, node)
		}
	}
}

func TestConsistentHashRemoveNode(t *testing.T) {
	ch := NewConsistentHash(100)

	ch.AddNode("node1")
	ch.AddNode("node2")

	key := "test-key"
	node1 := ch.GetNode(key)

	ch.RemoveNode(node1)
	node2 := ch.GetNode(key)

	if node1 == node2 {
		t.Error("Expected different node after removal")
	}
}
