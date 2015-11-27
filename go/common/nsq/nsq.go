package nsq

import (
	"strings"

	"github.com/bitly/go-nsq"

	"sirendaou.com/duserver/common/errors"
	sync_ "sirendaou.com/duserver/common/sync"
	"sirendaou.com/duserver/common/syslog"
)

const (
	MSG_CHAN_SIZE           = 10000
	PRODUCER_COUNT_PER_ADDR = 1
	CONSUMER_COUNT_PER_ADDR = 3
)

type nsqdMsg struct {
	topic string
	body  []byte
}
type nsqdMgr struct {
	addrs     string
	waitGroup *sync_.WaitGroup
	producers []*nsq.Producer
	msgCh     chan nsqdMsg
	consumers *ConsumerT
}

var g_nsqdMgr *nsqdMgr = nil

func Init(addrs string) {
	config := nsq.NewConfig()
	config.DefaultRequeueDelay = 0

	addrSlice := strings.Split(addrs, ",")
	produceCount := len(addrSlice) * PRODUCER_COUNT_PER_ADDR
	producers := make([]*nsq.Producer, produceCount)

	var err error = nil
	var idx int = 0
	for _, addr := range addrSlice {
		for i := 0; i < PRODUCER_COUNT_PER_ADDR; i++ {
			producers[idx], err = nsq.NewProducer(addr, config)
			if err != nil {
				panic(err)
			}
			idx++
		}
	}

	g_nsqdMgr = &nsqdMgr{
		addrs:     addrs,
		waitGroup: sync_.NewWaitGroup(),
		producers: producers,
		msgCh:     make(chan nsqdMsg, MSG_CHAN_SIZE),
		consumers: nil,
	}

	for i := 0; i < produceCount; i++ {
		go g_nsqdMgr.producer(i)
	}
	return
}

func (this *nsqdMgr) producer(idx int) {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()
	exitNotify := this.waitGroup.ExitNotify()

	for {
		select {
		case <-exitNotify:
			return
		case msg := <-this.msgCh:
			if err := this.producers[idx].Publish(msg.topic, msg.body); err != nil {
				syslog.Warn("Publish failed!!! ", err, msg.topic, string(msg.body))
				//投递失败，返回队列让其它链接投递
				this.msgCh <- msg
			}
		}
	}
}

func Publish(topic string, body []byte) {
	msg := nsqdMsg{
		topic: topic,
		body:  body,
	}
	g_nsqdMgr.msgCh <- msg
}

// message适配代码
type Message struct {
	*nsq.Message
}
type Handler interface {
	HandleMessage(message *Message) error
}
type ConsumerT struct {
	consumers []*nsq.Consumer
	handle    Handler
	msgCh     chan *Message
	waitGroup *sync_.WaitGroup
}

func (this *ConsumerT) HandleMessage(message *nsq.Message) error {
	this.msgCh <- &Message{message}
	return nil
}

func Consumer(topic, channel string, handle Handler) (*ConsumerT, error) {
	return ConsumerGO(topic, channel, 1, handle)
}

func ConsumerGO(topic, channel string, goCount uint, handle Handler) (*ConsumerT, error) {
	msgHandle := &ConsumerT{
		consumers: []*nsq.Consumer{},
		handle:    handle,
		msgCh:     make(chan *Message, MSG_CHAN_SIZE),
		waitGroup: sync_.NewWaitGroup(),
	}
	addrSlice := strings.Split(g_nsqdMgr.addrs, ",")
	for _, addr := range addrSlice {
		for i := 0; i < CONSUMER_COUNT_PER_ADDR; i++ {
			consumer, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
			if err != nil {
				return nil, errors.As(err, topic, channel)
			}
			consumer.SetLogger(nil, nsq.LogLevelInfo)
			consumer.AddHandler(msgHandle)
			if err := consumer.ConnectToNSQD(addr); err != nil {
				return nil, errors.As(err, topic, channel, g_nsqdMgr.addrs)
			}
			msgHandle.consumers = append(msgHandle.consumers, consumer)
		}
	}
	g_nsqdMgr.consumers = msgHandle
	for i := 0; i < int(goCount); i++ {
		go msgHandle.work()
	}
	return msgHandle, nil
}

func (this *ConsumerT) work() {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()
	exitNotify := this.waitGroup.ExitNotify()
	for {
		select {
		case <-exitNotify:
			return
		case msg := <-this.msgCh:
			if err := this.handle.HandleMessage(msg); err != nil {
				syslog.Info(err, msg)
				//this.msgCh <- msg
			}
		}
	}
}

func Deinit() {
	g_nsqdMgr.waitGroup.Wait()

	for _, p := range g_nsqdMgr.producers {
		p.Stop()
	}
	if g_nsqdMgr.consumers != nil {
		for _, c := range g_nsqdMgr.consumers.consumers {
			c.Stop()
		}
		g_nsqdMgr.consumers.waitGroup.Wait()
	}
}
