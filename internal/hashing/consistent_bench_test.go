package hashing

import (
	"testing"
)

func BenchmarkConsistentHash(b *testing.B) {
	ch := NewConsistentHash(100)
	ch.AddNode("node1")
	ch.AddNode("node2")
	ch.AddNode("node3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.GetNode("test-key")
	}
}
