package container

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	queue := NewQueue()

	queue.NormalIn(1)
	queue.LowIn(2)
	queue.HighIn(3)

	fmt.Println(queue.Out())
	fmt.Println(queue.Out())
	fmt.Println(queue.Out())
	fmt.Println(queue.Out())
}

func TestQueuePerformance(t *testing.T) {
	total := 1000000
	goCount := 100
	Count := total / goCount
	wp := &sync.WaitGroup{}
	queue := NewQueue()
	exit := false
	start_time := time.Now().Unix()
	for i := 0; i < goCount; i++ {
		go func() {
			fmt.Println("==============")
			wp.Add(1)
			defer wp.Done()
			for j := 0; j < Count; j++ {
				queue.NormalIn(j)
			}
		}()
	}

	for i := 0; i < goCount; i++ {
		go func() {
			for !exit {
				queue.Out()
			}
		}()
	}

	wp.Wait()
	for queue.Len() != 0 {
		time.Sleep(time.Millisecond * 100)
	}

	end_time := time.Now().Unix()

	fmt.Println("use time:", end_time-start_time, queue.Len())
}
