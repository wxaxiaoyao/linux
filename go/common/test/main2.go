package main

import (
	"encoding/json"
	"time"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/nsq"
)

type Msg struct {
	Id         int64
	Start_time int64
}

func (m *Msg) pack() []byte {
	body, err := json.Marshal(m)
	panic(err)
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
	time.Sleep(time.Millisecond * 3)
	common.NsqPublish("test_topic1", message.Body)
	return nil
}

func main() {
	common.NsqConsumer("test_topic2", "test_channel", handle{})
	//common.NsqConsumer("test_topic2", "test_channel", handle{})
	//common.NsqConsumer("test_topic2", "test_channel", handle{})
	select {}
}
