package db

type qNode struct {
	value string
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
	node := &qNode{value: item}
	q.Length++
	if q.Length == 1 {
		q.head, q.tail = node, node
		return
	}

	q.tail.next = node
	q.tail = node
}

func (q *Queue) Dequeue() string {
	if q.Length == 0 {
		return ""
	}

	q.Length--
	head := q.head
	q.head = q.head.next
	return head.value
}

func (q *Queue) Peek() string {
	if q.Length == 0 {
		return ""
	}

	return q.head.value
}
