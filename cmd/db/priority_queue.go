package db

import (
	"fmt"
	"time"
)

type pqItem struct {
	priority   int
	insertedAt int64
	data       string
}

func newPqItem(data string, priority int) pqItem {
	return pqItem{data: data, priority: priority, insertedAt: time.Now().UnixMilli()}
}

// Check if an item is of lower priority from another, it first checks the priority
// and if those are equal it checks the time created.
func (item *pqItem) isLowerThan(other pqItem) bool {
	return item.priority < other.priority || item.insertedAt < other.insertedAt
}

// Check if an item is of higher priority from another, it first checks the priority
// and if those are equal it checks the time created.
func (item *pqItem) isHigherThan(other pqItem) bool {
	return item.priority >= other.priority || item.insertedAt >= other.insertedAt
}

// This Memo data structure has no Redis equivalent, it is an implementation of a
// priority queue based on a Min Heap. A lower priority means that an item will be retrieved
// first (eg. an item with priority 1 with be retrieved before an item with priority 2).
// If the priorities are the same then the item which was creted first will be retrieved.
type PriorityQueue struct {
	Length int
	items  []pqItem
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

func (p *PriorityQueue) Debug() {
	for i, v := range p.items {
		fmt.Printf("%d  { v: %s, p: %d, t: %d } \n", i, v.data, v.priority, v.insertedAt)
	}
}

func (p *PriorityQueue) Enqueue(data string, priority int) {
	p.items = append(p.items, newPqItem(data, priority))
	p.heapifyUp(p.Length)
	p.Length++
}

func (p *PriorityQueue) Dequeue() string {
	if p.Length == 0 {
		return ""
	}

	out := p.items[0]
	p.Length--
	if p.Length == 0 {
		p.items = []pqItem{}
		return out.data
	}

	p.items[0] = p.items[p.Length]
	p.heapifyDown(0)
	return out.data
}

func (p *PriorityQueue) Peek() string {
	if p.Length == 0 {
		return ""
	}

	return p.items[0].data
}

// Move up to the correct position in the heap
func (p *PriorityQueue) heapifyUp(idx int) {
	if idx == 0 {
		return
	}

	pIdx := p.parent(idx)
	parent := p.items[pIdx]
	current := p.items[idx]
	if current.isLowerThan(parent) {
		p.swapItems(idx, pIdx)
		p.heapifyUp(pIdx)
	}
}

// Move down to the correct position in the heap
func (p *PriorityQueue) heapifyDown(idx int) {
	if idx >= p.Length {
		return
	}

	lIdx := p.leftChild(idx)
	rIdx := p.rightChild(idx)
	if lIdx >= p.Length {
		return
	}

	current := p.items[idx]
	left := p.items[lIdx]
	right := p.items[rIdx]

	if right.isLowerThan(left) && right.isLowerThan(current) {
		p.swapItems(idx, rIdx)
		p.heapifyDown(rIdx)
	} else if right.isHigherThan(left) && left.isLowerThan(current) {
		p.swapItems(idx, lIdx)
		p.heapifyDown(lIdx)
	}
}

func (p *PriorityQueue) swapItems(i int, j int) {
	p.items[i], p.items[j] = p.items[j], p.items[i]
}

func (p *PriorityQueue) parent(idx int) int {
	return (idx - 1) / 2
}

func (p *PriorityQueue) leftChild(idx int) int {
	return (idx * 2) + 1
}

func (p *PriorityQueue) rightChild(idx int) int {
	return (idx * 2) + 2
}
