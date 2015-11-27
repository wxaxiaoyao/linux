package sync

import (
	"fmt"
	"testing"
	"time"
)

func routine(wp *WaitGroup, i int) {
	wp.Add(1)
exitLabel:
	for {
		select {
		case <-wp.ExitNotify():
			fmt.Println("exit:", i)
			break exitLabel
		}
		fmt.Println("test:", i)
	}
	fmt.Println("Sleep:", i)
	time.Sleep(time.Second)
	wp.Done()
	println("wait finish:", i)
}

func TestWaitGroup(t *testing.T) {
	wp := NewWaitGroup()

	for i := 0; i < 3; i++ {
		go routine(wp, i)
	}

	go func() {
		wp.Wait()
		fmt.Println("wait exit")
	}()

	go routine(wp, 3)

	time.Sleep(time.Second * 10)
}
