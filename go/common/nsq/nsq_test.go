package nsq

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

var g_timeMap map[int64]int64 = map[int64]int64{}
var max_time int64 = 0
var count = 0
var lock = &sync.Mutex{}
var lock1 = &sync.Mutex{}

type handle1 struct {
}

func (handle1) HandleMessage(message *Message) error {
	Publish("test_topic2", message.Body)
	//log.Println("handle1:", string(message.Body))
	return nil
}

type handle2 struct {
}

func (handle2) HandleMessage(message *Message) error {
	id, _ := strconv.ParseInt(string(message.Body), 10, 32)
	lock.Lock()
	start_time := g_timeMap[id]
	lock.Unlock()
	end_time := time.Now().Unix()
	use_time := end_time - start_time
	if use_time > max_time {
		max_time = use_time
		fmt.Printf("id:%v, start_time:%v, end_time:%v, use_time:%v\n", id, start_time, end_time, use_time)
	}
	lock1.Lock()
	count++
	lock1.Unlock()
	if count%100000 == 0 {
		log.Println("=============================")
	}
	//log.Println("handle2:", count)
	return nil
}

func TestNsqMgr(t *testing.T) {
	Init("127.0.0.1:4150")
	//Init("127.0.0.1:4150,127.0.0.1:4148,127.0.0.1:4152,127.0.0.1:4154,127.0.0.1:4156,127.0.0.1:4158")
	max_count := 1000000
	//start_time := time.Now()
	ConsumerGO("test_topic1", "test_channel", 100, handle1{})
	ConsumerGO("test_topic2", "test_channel", 100, handle2{})

	start_time := time.Now().Unix()
	var t1, t2 int64 = 0, 0
	for i := 0; i < max_count; i++ {
		id := int64(1000000 + i)
		lock.Lock()
		g_timeMap[id] = time.Now().Unix()
		lock.Unlock()
		Publish("test_topic1", []byte(fmt.Sprint(id)))
		if i%100000 == 0 {
			t1 = t2
			t2 = time.Now().Unix()
			fmt.Println("100000:", t2-t1)
		}
		//time.Sleep(time.Millisecond)
	}
	end_time := time.Now().Unix()
	fmt.Println("Pushlish Time:", end_time-start_time)
	for count < max_count {
	}
	end_time = time.Now().Unix()

	fmt.Println("use_time:", end_time-start_time)
	time.Sleep(time.Second)
	Deinit()
}

type handle struct {
	count int
	t1    int64
	t2    int64
}

func (h *handle) HandleMessage(message *Message) error {
	if h.count%100000 == 0 {
		h.t1 = h.t2
		h.t2 = time.Now().Unix()
		fmt.Println("100000 use time:", h.t2-h.t1)
	}
	h.count++
	return nil
}

func TestNsqPC(t *testing.T) {
	Init("127.0.0.1:4150,127.0.0.1:4148,127.0.0.1:4152,127.0.0.1:4154,127.0.0.1:4156,127.0.0.1:4158")
	if _, err := Consumer("test_topic", "test_channel", &handle{}); err != nil {
		fmt.Println(err)
		return
	}
	Publish("test_topic", []byte("hello world"))
	time.Sleep(time.Second * 2)
	Deinit()
	time.Sleep(time.Second * 2)
}

func TestNsqProduce(t *testing.T) {
	//Init("127.0.0.1:4150")
	Init("127.0.0.1:4150,127.0.0.1:4148,127.0.0.1:4152,127.0.0.1:4154,127.0.0.1:4156,127.0.0.1:4158")
	wg := &sync.WaitGroup{}
	max_count := 100000
	start_time := time.Now().Unix()
	for j := 0; j < 20; j++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			for i := 0; i < max_count; i++ {
				Publish("test_topic", []byte("hello world"))
			}
		}()
	}
	time.Sleep(time.Second)
	wg.Wait()
	end_time := time.Now().Unix()
	fmt.Println("Pushlish Time:", end_time-start_time)
	time.Sleep(time.Second * 5)
	Deinit()
}

func TestNsqConsume(t *testing.T) {
	Init("127.0.0.1:4150")
	//Init("127.0.0.1:4150,127.0.0.1:4148,127.0.0.1:4152,127.0.0.1:4154,127.0.0.1:4156,127.0.0.1:4158")
	if _, err := Consumer("test_topic", "test_channel", &handle{}); err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(time.Minute * 5)
	Deinit()
}
