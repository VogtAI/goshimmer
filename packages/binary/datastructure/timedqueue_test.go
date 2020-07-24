package datastructure

import (
	"fmt"
	"testing"
	"time"
)

func TestTimedQueue_Poll(t *testing.T) {
	timedQueue := NewTimedQueue()

	go func() {
		for timedQueue.IsProcessingElements(true) {
			for currentEntry := timedQueue.Poll(); currentEntry != nil; currentEntry = timedQueue.Poll() {
				fmt.Println(currentEntry)
			}
		}
	}()

	timedQueue.Add(2, time.Now().Add(1*time.Second))
	elem := timedQueue.Add(4, time.Now().Add(2*time.Second))
	timedQueue.Add(6, time.Now().Add(3*time.Second))

	time.Sleep(4 * time.Second)

	elem.Cancel()

	timedQueue.Add(6, time.Now().Add(time.Second))
	timedQueue.Add(68, time.Now().Add(4*time.Second))

	timedQueue.Shutdown(true)

	time.Sleep(500 * time.Millisecond)
}
