package sync

import (
	"sync"
)

type WaitGroup struct {
	exitCh   chan bool       //子协程是否应结束，主协程退出通知
	exit     bool            //主协程退出标志
	waitGrop *sync.WaitGroup //等待子协程退出
	count    int             //子协程的数量
	mutex    *sync.Mutex     //count互斥访问
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		exitCh:   make(chan bool, 1), // 缓存一个用于自己给自己发送，不会死锁
		waitGrop: &sync.WaitGroup{},
		count:    0,
		mutex:    &sync.Mutex{},
		exit:     false,
	}
}

func (wp *WaitGroup) Add(delta int) {
	wp.waitGrop.Add(delta)
	wp.mutex.Lock()
	if wp.exit {
		//退出信号已发出，故自己发给自己并不计数
		wp.exitCh <- wp.exit
	} else {
		wp.count += delta
	}
	wp.mutex.Unlock()
}

func (wp *WaitGroup) AddOne() {
	wp.Add(1)
}

func (wp *WaitGroup) Done() {
	wp.mutex.Lock()
	wp.count += (-1)
	wp.mutex.Unlock()
	wp.waitGrop.Done()
}

func (wp *WaitGroup) ExitNotify() <-chan bool {
	return wp.exitCh
}

func (wp *WaitGroup) Wait() {
	wp.mutex.Lock()
	count := wp.count
	//wp.exit屏蔽：调用wait后，可再调用add，此时协程会阻塞，等待wait发退出通知         用于重覆 易出错
	//wp.exit不屏蔽：调用wait后，可再调用add，此时协程不会阻塞，自己给自己发退出通知   用一次性 较安全
	//wp.exit = true
	wp.mutex.Unlock()

	for i := 0; i < count; i++ {
		wp.exitCh <- wp.exit
	}
	wp.waitGrop.Wait()
}
