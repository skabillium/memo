package db

import "testing"

func TestQueue(t *testing.T) {
	q := NewQueue()
	q.Enqueue("5")
	q.Enqueue("7")
	q.Enqueue("9")

	if q.Length != 3 {
		t.Error("Expected Length to be 3")
	}
	if q.Dequeue() != "5" {
		t.Error("Expected Dequeue() to return 5")
	}
	if q.Dequeue() != "7" {
		t.Error("Expected Dequeue() to return 7")
	}

	if q.Peek() != "9" {
		t.Error("Expected Peek() to return 9")
	}
	if q.Length != 1 {
		t.Error("Expected Length to be 1")
	}
	if q.Dequeue() != "9" {
		t.Error("Expected Dequeue() to return 9")
	}
	if q.Length != 0 {
		t.Error("Expected Length to be 0")
	}

	q.Enqueue("11")
	if q.Length != 1 {
		t.Error("Expected Length to be 1")
	}
	if q.Dequeue() != "11" {
		t.Error("Expected Dequeue() to return 11")
	}
	if q.Dequeue() != "" {
		t.Error("Expected Dequeue() to return nil")
	}
	if q.Peek() != "" {
		t.Error("Expected Peek to return nil")
	}
	if q.Length != 0 {
		t.Error("Expected Length to be 0")
	}
}
