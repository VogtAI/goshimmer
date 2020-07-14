package datastructure

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerQueue_HasNext(t *testing.T) {
	tq := NewTimerQueue()

	tq.Schedule(1, 2, time.Now().Add(5*time.Second))

	for tq.HasNext() {
		fmt.Println(tq.Next())

		break
	}
}
