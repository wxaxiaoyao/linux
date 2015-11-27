package main

import (
	"encoding/json"
	"fmt"
	"time"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/nsq"
)

var max_time int64 = 0

type Msg struct {
	Id         int64
	Start_time int64
}

func (m *Msg) pack() []byte {
	body, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return body
}

func (m *Msg) unpack(body []byte) {
	if err := json.Unmarshal(body, m); err != nil {
		panic(err)
	}
}

type handle struct {
}

func (handle) HandleMessage(message *nsq.Message) error {
	m := &Msg{}
	m.unpack(message.Body)
	start_time := m.Start_time
	end_time := time.Now().Unix()
	use_time := end_time - start_time
	if use_time > max_time {
		max_time = use_time
		fmt.Printf("id:%v, start_time:%v, end_time:%v, use_time:%v\n", m.Id, start_time, end_time, use_time)
	}
	return nil
}
func main() {
	max_count := 30000
	common.NsqConsumer("test_topic1", "test_channel", handle{})

	m := &Msg{}
	for i := 0; i < max_count; i++ {
		m.Id = int64(1000000 + i)
		m.Start_time = time.Now().Unix()
		common.NsqPublish("test_topic2", m.pack())
		time.Sleep(time.Millisecond)
	}

	fmt.Println("finish")
	select {}
}
