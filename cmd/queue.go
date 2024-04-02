package main

type qNode struct {
	value MemoString
	next  *qNode
}

type Queue struct {
	Length int
	head   *qNode
	tail   *qNode
}

func NewQueue() *Queue {
	return &Queue{Length: 0, head: nil, tail: nil}
}

func (q *Queue) Enqueue(item string) {
	node := &qNode{value: MemoString(item)}
	q.Length++
	if q.Length == 1 {
		q.head, q.tail = node, node
		return
	}

	q.tail.next = node
	q.tail = node
}

func (q *Queue) Dequeue() MemoString {
	if q.Length == 0 {
		return ""
	}

	q.Length--
	head := q.head
	q.head = q.head.next
	return head.value
}

func (q *Queue) Peek() MemoString {
	if q.Length == 0 {
		return ""
	}

	return q.head.value
}
