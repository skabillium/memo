package main

type lNode struct {
	value string
	next  *lNode
	prev  *lNode
}

type List struct {
	Length int
	head   *lNode
	tail   *lNode
}

func NewList() *List {
	return &List{Length: 0, head: nil, tail: nil}
}

func (l *List) Prepend(item string) {
	l.Length++
	node := &lNode{value: item}
	if l.Length == 1 {
		l.head, l.tail = node, node
		return
	}

	node.next = l.head
	l.head.prev = node
	l.head = node
}

func (l *List) Append(item string) {
	if l.Length == 0 {
		l.Prepend(item)
		return
	}

	l.Length++
	node := &lNode{value: item}
	l.tail.next = node
	node.prev = l.tail
	l.tail = node
}

func (l *List) PopHead() string {
	if l.Length == 0 {
		return ""
	}

	l.Length--
	head := l.head
	l.head = head.next
	return head.value
}

func (l *List) PopTail() string {
	if l.Length == 0 {
		return ""
	}

	l.Length--
	tail := l.tail
	l.tail = l.tail.prev
	return tail.value
}

func (l *List) PeekHead() string {
	if l.Length == 0 {
		return ""
	}
	return l.head.value
}

func (l *List) PeekTail() string {
	if l.Length == 0 {
		return ""
	}
	return l.tail.value
}
