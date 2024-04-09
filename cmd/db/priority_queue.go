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
	return pqItem{data: data, priority: priority, insertedAt: time.Now().Unix()}
}

func (item *pqItem) isLowerThan(other pqItem) bool {
	return item.priority < other.priority || item.insertedAt < other.insertedAt
}

func (item *pqItem) isHigherThan(other pqItem) bool {
	return item.priority >= other.priority || item.insertedAt >= other.insertedAt
}

type PriorityQueue struct {
	Length int
	items  []pqItem
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

func (p *PriorityQueue) Debug() {
	for i, v := range p.items {
		fmt.Printf("%d  { v: %s, p: %d } \n", i, v.data, v.priority)
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

func (p *PriorityQueue) heapifyDown(idx int) {
	if idx >= p.Length {
		return
	}

	lIdx := p.leftChild(idx)
	rIdx := p.rightChild(idx)
	if lIdx >= p.Length {
		return
	}

	// priority := p.items[idx].priority
	// leftPr := p.items[lIdx].priority
	// rightPr := p.items[rIdx].priority

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
	return

	tmpd := p.items[i].data
	tmpp := p.items[i].priority
	tmpi := p.items[i].insertedAt

	p.items[i].data = p.items[j].data
	p.items[i].priority = p.items[j].priority
	p.items[i].insertedAt = p.items[j].insertedAt

	p.items[j].data = tmpd
	p.items[j].priority = tmpp
	p.items[j].insertedAt = tmpi
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
