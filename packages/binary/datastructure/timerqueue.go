package datastructure

import (
	"container/heap"
	"sync"
	"time"
)

type TimerQueue struct {
	directory map[interface{}]*timerHeapElement
	heap      timerHeap
	heapMutex sync.RWMutex
}

func NewTimerQueue() *TimerQueue {
	return &TimerQueue{
		directory: make(map[interface{}]*timerHeapElement),
	}
}

func (t *TimerQueue) Schedule(identifier interface{}, value interface{}, scheduledTime time.Time) {
	t.heapMutex.Lock()
	defer t.heapMutex.Unlock()

	heapElement, elementQueued := t.directory[identifier]
	if elementQueued {
		heapElement.time = scheduledTime
		heap.Fix(&t.heap, heapElement.index)

		return
	}

	heapElement = &timerHeapElement{
		identifier: identifier,
		value:      value,
		time:       scheduledTime,
		index:      0,
	}

	heap.Push(&t.heap, heapElement)
	t.directory[identifier] = heapElement
}

func (t *TimerQueue) Size() int {
	t.heapMutex.RLock()
	defer t.heapMutex.RUnlock()

	return len(t.heap)
}

func (t *TimerQueue) HasNext() bool {
	return t.Size() != 0
}

func (t *TimerQueue) Next() interface{} {
	t.heapMutex.RLock()
	firstElement := t.heap[0]
	t.heapMutex.RUnlock()

	<-time.After(firstElement.time.Sub(time.Now()))

	return firstElement.value
}

type timerHeapElement struct {
	identifier interface{}
	value      interface{}
	time       time.Time
	index      int
}

type timerHeap []*timerHeapElement

func (h timerHeap) Len() int {
	return len(h)
}

func (h timerHeap) Less(i, j int) bool {
	return h[i].time.Before(h[j].time)
}

func (h timerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

func (h *timerHeap) Push(x interface{}) {
	data := x.(*timerHeapElement)
	*h = append(*h, data)
	data.index = len(*h) - 1
}

func (h *timerHeap) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	*h = (*h)[:n-1]
	data.index = -1
	return data
}

var _ heap.Interface = &timerHeap{}
