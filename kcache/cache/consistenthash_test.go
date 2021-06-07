package cache

import (
	"log"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	ch := NewConsistentHash(5, nil)
	ch.Add("aaa", "bbb", "ccc")
	log.Printf("[hash include]:%v", ch.keys)
	if len(ch.keys) != 15 {
		t.Fatalf("Add error, true num: %d", len(ch.keys))
	}

	ch.Delete("aaa", "bbb", "ccc")
	log.Printf("[hash include]:%v", ch.keys)
	if len(ch.keys) != 0 {
		t.Fatalf("Delete error, true num: %d", len(ch.keys))
	}
}
