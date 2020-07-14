package datastructure

import (
	"fmt"
	"testing"
	"time"
)

func TestTimedQueue_Poll(t *testing.T) {
	tq := NewTimedQueue()

	go func() {
		for tq.WaitForNewElements() {
			for currentEntry := tq.Poll(); currentEntry != nil; currentEntry = tq.Poll() {
				fmt.Println(currentEntry)
			}
		}
	}()

	tq.Add(2, time.Now().Add(1*time.Second))
	elem := tq.Add(4, time.Now().Add(2*time.Second))
	tq.Add(6, time.Now().Add(3*time.Second))

	elem.Cancel()

	time.Sleep(3 * time.Second)

	tq.Add(6, time.Now().Add(time.Second))
	tq.Add(68, time.Now().Add(4*time.Second))

	tq.Shutdown(false)

	time.Sleep(500 * time.Millisecond)
}
