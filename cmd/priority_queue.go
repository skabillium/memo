package main

type pqItem struct {
	priority int
	data     MemoString
}

type PriorityQueue struct {
	Length int
	items  []pqItem
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

func (p *PriorityQueue) Enqueue(data string, priority int) {
	p.items = append(p.items, pqItem{priority: priority, data: MemoString(data)})
	p.heapifyUp(p.Length)
	p.Length++
}

func (p *PriorityQueue) Dequeue() MemoString {
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

func (p *PriorityQueue) Peek() MemoString {
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
	parentPr := p.items[pIdx].priority
	priority := p.items[idx].priority
	if parentPr > priority {
		// Swap values
		tmp := p.items[pIdx].data
		p.items[pIdx].priority = priority
		p.items[pIdx].data = p.items[idx].data

		p.items[idx].priority = parentPr
		p.items[idx].data = tmp
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

	priority := p.items[idx].priority
	leftPr := p.items[lIdx].priority
	rightPr := p.items[rIdx].priority

	if leftPr > rightPr && priority > rightPr {
		p.swapItems(idx, rIdx)
		p.heapifyDown(rIdx)
	} else if rightPr > leftPr && priority > leftPr {
		p.swapItems(idx, lIdx)
		p.heapifyDown(lIdx)
	}
}

func (p *PriorityQueue) swapItems(i int, j int) {
	tmpd := p.items[i].data
	tmpp := p.items[i].priority

	p.items[i].data = p.items[j].data
	p.items[i].priority = p.items[j].priority
	p.items[j].data = tmpd
	p.items[j].priority = tmpp
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
