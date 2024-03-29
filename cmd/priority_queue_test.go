package main

import "testing"

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	pq.Enqueue("2", 2)
	pq.Enqueue("1", 1)
	pq.Enqueue("3", 1)

	if pq.Length != 3 {
		t.Error("Expected Length to be 3")
	}
	if pq.Dequeue() != "1" {
		t.Error("Expected Dequeue() to return 1")
	}
	if pq.Dequeue() != "3" {
		t.Error("Expected Dequeue() to return 3")
	}
	if pq.Peek() != "2" {
		t.Error("Expected Peek() to return 2")
	}
	if pq.Dequeue() != "2" {
		t.Error("Expected Dequeue() to return 2")
	}
	if pq.Length != 0 {
		t.Error("Expected Length to be 0")
	}
	if pq.Dequeue() != "" {
		t.Error("Expected Dequeue() to return ''")
	}
	if pq.Peek() != "" {
		t.Error("Expected Peek() to return ''")
	}
}
