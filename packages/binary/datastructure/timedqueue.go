package datastructure

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/typeutils"
)

// region TimedQueue ///////////////////////////////////////////////////////////////////////////////////////////////////

// TimedQueue represents a queue, that holds values that will only be released at a given time. The corresponding Poll
// method waits for the element to be available before it returns its value and is therefore blocking.
type TimedQueue struct {
	heap                               timedHeap
	heapMutex                          sync.RWMutex
	nonEmpty                           sync.WaitGroup
	shutdown                           chan byte
	returnPendingElementsAfterShutdown typeutils.AtomicBool
}

// NewTimedQueue is the constructor for the TimedQueue.
func NewTimedQueue() (queue *TimedQueue) {
	queue = &TimedQueue{
		shutdown: make(chan byte),
	}
	queue.nonEmpty.Add(1)

	return
}

// Add inserts a new element into the queue that can be retrieved via Poll() at the specified time.
func (t *TimedQueue) Add(value interface{}, scheduledTime time.Time) (addedElement *TimedQueueElement) {
	// sanitize parameters
	if value == nil {
		panic("<nil> must not be added to the queue")
	}

	// acquire locks
	t.heapMutex.Lock()
	defer t.heapMutex.Unlock()

	// mark queue as non-empty
	if len(t.heap) == 0 {
		t.nonEmpty.Done()
	}

	// add new element
	addedElement = &TimedQueueElement{
		timedQueue: t,
		value:      value,
		time:       scheduledTime,
		cancel:     make(chan byte),
		index:      0,
	}
	heap.Push(&t.heap, addedElement)

	return
}

// Poll returns the first value of this queue. It waits for the scheduled time before returning and is therefore
// blocking. It returns nil if the queue is empty.
func (t *TimedQueue) Poll() interface{} {
	// acquire locks and abort if empty
	t.heapMutex.Lock()
	if len(t.heap) == 0 {
		t.heapMutex.Unlock()

		return nil
	}

	// retrieve first element
	polledElement := heap.Remove(&t.heap, 0).(*TimedQueueElement)

	// mark the queue as empty if last element was polled
	if len(t.heap) == 0 {
		t.nonEmpty.Add(1)
	}
	t.heapMutex.Unlock()

	// wait for the return value to become due
	select {
	// abort if the queue was shutdown
	case <-t.shutdown:
		if !t.returnPendingElementsAfterShutdown.IsSet() {
			if t.Size() != 0 {
				return t.Poll()
			}

			return nil
		}

		fmt.Println("RETURN LAST ELEMENTS")

		return polledElement.value
	// abort waiting for this element and wait for the next one instead if it was canceled
	case <-polledElement.cancel:
		return t.Poll()

	// return the result after the time is reached
	case <-time.After(time.Until(polledElement.time)):
		return polledElement.value
	}
}

// Size returns the amount of elements that are currently enqueued in this queue.
func (t *TimedQueue) Size() int {
	t.heapMutex.RLock()
	defer t.heapMutex.RUnlock()

	return len(t.heap)
}

// WaitForNewElements waits for the queue to be non-empty. This can be used by i.e. schedulers to continuously iterate
// over the queue to process its elements. It returns false, if the queue has been shutdown.
func (t *TimedQueue) WaitForNewElements() bool {
	t.nonEmpty.Wait()

	select {
	case <-t.shutdown:
		return t.returnPendingElementsAfterShutdown.IsSet() && t.Size() != 0
	default:
		return true
	}
}

// Shutdown terminates the queue and stops the currently waiting Poll requests.
func (t *TimedQueue) Shutdown(returnPendingElements bool) {
	t.heapMutex.Lock()
	select {
	// only shutdown once
	case <-t.shutdown:
		// do nothing
	default:
		t.returnPendingElementsAfterShutdown.SetTo(returnPendingElements)

		close(t.shutdown)
	}
	t.heapMutex.Unlock()

	t.nonEmpty.Wait()
}

// removeElement is an internal utility function that removes the given element from the queue.
func (t *TimedQueue) removeElement(element *TimedQueueElement) {
	// abort if the element was removed already
	if element.index == -1 {
		return
	}

	// remove the element
	heap.Remove(&t.heap, element.index)

	// mark the queue as empty
	if len(t.heap) == 0 {
		t.nonEmpty.Add(1)
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TimedQueueElement ////////////////////////////////////////////////////////////////////////////////////////////

// TimedQueueElement is an element in the TimedQueue. It
type TimedQueueElement struct {
	timedQueue *TimedQueue
	value      interface{}
	cancel     chan byte
	time       time.Time
	index      int
}

// Cancel removed the given element from the queue and cancels its execution.
func (timedQueueElement *TimedQueueElement) Cancel() {
	// acquire locks
	timedQueueElement.timedQueue.heapMutex.Lock()
	defer timedQueueElement.timedQueue.heapMutex.Unlock()

	// remove element from queue
	timedQueueElement.timedQueue.removeElement(timedQueueElement)

	// close the cancel channel to notify subscribers
	close(timedQueueElement.cancel)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region timedHeap ////////////////////////////////////////////////////////////////////////////////////////////////////

// timedHeap defines a heap based on times.
type timedHeap []*TimedQueueElement

// Len is the number of elements in the collection.
func (h timedHeap) Len() int {
	return len(h)
}

// Less reports whether the element with index i should sort before the element with index j.
func (h timedHeap) Less(i, j int) bool {
	return h[i].time.Before(h[j].time)
}

// Swap swaps the elements with indexes i and j.
func (h timedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

// Push adds x as the last element to the heap.
func (h *timedHeap) Push(x interface{}) {
	data := x.(*TimedQueueElement)
	*h = append(*h, data)
	data.index = len(*h) - 1
}

// Pop removes and returns the last element of the heap.
func (h *timedHeap) Pop() interface{} {
	n := len(*h)
	data := (*h)[n-1]
	*h = (*h)[:n-1]
	data.index = -1
	return data
}

// interface contract (allow the compiler to check if the implementation has all of the required methods).
var _ heap.Interface = &timedHeap{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
