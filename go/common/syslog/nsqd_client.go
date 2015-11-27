package syslog

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bitly/go-nsq"
	sync_ "sirendaou.com/duserver/common/sync"
)

type nsqdLogMgr struct {
	waitGroup     *sync_.WaitGroup
	logTopic      string
	producers     []*nsq.Producer
	producerCount uint
	logMsgCh      chan *LogMsg
	logLevel      int
	init          bool
}

func init() {
	RegisterLogger(NSQD_CLIENT, &nsqdLogMgr{})
}

func (l *nsqdLogMgr) Init(config *Config) error {
	if config.NsqdLog == false {
		return errors.New("nsqd log server not configure!!!")
	}

	nsqConfig := nsq.NewConfig()
	nsqConfig.DefaultRequeueDelay = 0
	connCountPerAddr := 2
	addrSlice := strings.Split(config.NsqdAddrs, ",")
	producerCount := len(addrSlice) * connCountPerAddr
	producers := make([]*nsq.Producer, producerCount)

	var err error
	idx := 0
	for _, addr := range addrSlice {
		for i := 0; i < connCountPerAddr; i++ {
			producers[idx], err = nsq.NewProducer(addr, nsqConfig)
			if err != nil {
				fmt.Println("NewProducer ", addr, " error:", err)
			}
			//fmt.Println("nsq_client connect success:", i)
			idx++
		}
	}

	l.waitGroup = sync_.NewWaitGroup()
	l.producers = producers
	l.producerCount = uint(producerCount)
	l.logMsgCh = make(chan *LogMsg, getChildLogScaleSize())
	l.logTopic = config.NsqdTopic
	l.logLevel = config.NsqdLogLevel
	l.init = true

	for i := 0; i < producerCount; i++ {
		go l.run(i)
	}
	return nil
}

func (l *nsqdLogMgr) run(idx int) {
	l.waitGroup.Add(1)
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()

	//fmt.Println("nsqd_client log server start...")
	for {
		select {
		case <-exitNotify:
			//fmt.Println("nsqd_client log server stop!!!")
			return
		case msg := <-l.logMsgCh:
			body, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("Json.Marshal failed:", err)
				continue
			}
			if err := l.producers[idx].Publish(l.logTopic, body); err != nil {
				fmt.Println("nsq.Publish error:", err)
				continue
			}
		}
	}
}

func (l *nsqdLogMgr) Write(logMsg *LogMsg) {
	if l.logLevel <= logMsg.Level {
		l.logMsgCh <- logMsg
		/*
			select {
			case l.logMsgCh <- logMsg:
			default:
				fmt.Println("===================msg too many======================")
			}
		*/
	}
}

func (l *nsqdLogMgr) IsValid() bool {
	return l.init
}

func (l *nsqdLogMgr) Deinit() {
	if l.init == false {
		return
	}
	l.waitGroup.Wait()

	for _, produce := range l.producers {
		produce.Stop()
	}
	l.init = false
}
