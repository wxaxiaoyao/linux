package syslog

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sync_ "sirendaou.com/duserver/common/sync"

	"github.com/bitly/go-nsq"
)

var (
	CONSUME_COUNT_PER_ADDR = 5
)

type nsqdSvrLogMgr struct {
	waitGroup *sync_.WaitGroup
	logLevel  int
	consumer  []*nsq.Consumer
	// 日志输出
	logger map[string]Logger
	init   bool
}

func init() {
	RegisterLogger(NSQD_SERVER, &nsqdSvrLogMgr{})
}

func (l *nsqdSvrLogMgr) Init(config *Config) error {
	// 一个模块只创建一个实例
	if config.NsqdSvrLog == false {
		return errors.New("nsqd_server log server not configure!!!")
	}
	// 注册日志驱动
	l.logLevel = config.LogLevel
	l.logger = make(map[string]Logger)
	l.waitGroup = sync_.NewWaitGroup()

	nsqdCfg := config.NsqdSvrConfig
	if nsqdCfg.ConsoleLog {
		l.logger[CONSOLE] = &consoleLogMgr{}
	}
	if nsqdCfg.FileLog {
		l.logger[FILE] = &fileLogMgr{}
	}
	if nsqdCfg.NsqdLog {
		l.logger[NSQD_CLIENT] = &nsqdLogMgr{}
	}
	if nsqdCfg.NsqdSvrLog {
		l.logger[NSQD_SERVER] = &nsqdSvrLogMgr{}
	}
	if nsqdCfg.MysqlLog {
		l.logger[MYSQL] = &MysqlLogMgr{}
	}
	if nsqdCfg.MongoLog {
		l.logger[MONGO] = &MongoLogMgr{}
	}
	// 初始化日志驱动
	for logName, logger := range l.logger {
		fmt.Printf("init log driver...[%s] [%s]\n", logName, nsqdCfg.ModelName)
		if err := logger.Init(nsqdCfg); err != nil {
			fmt.Printf("init log driver failed!!![%s] [%s]; error:%s\n", logName, nsqdCfg.ModelName, err.Error())
			continue
		}
		fmt.Printf("init log driver success...[%s] [%s]\n", logName, nsqdCfg.ModelName)
	}
	l.consumer = []*nsq.Consumer{}
	addrs := strings.Split(config.NsqdAddrs, ",")
	for _, addr := range addrs {
		for i := 0; i < CONSUME_COUNT_PER_ADDR; i++ {
			consumer, err := nsq.NewConsumer(config.NsqdTopic, config.NsqdChannel, nsq.NewConfig())
			if err != nil {
				fmt.Println("NewConsumer failed:", err)
				return err
			}
			consumer.SetLogger(nil, nsq.LogLevelInfo)
			consumer.AddHandler(l)
			if err := consumer.ConnectToNSQD(addr); err != nil {
				fmt.Println("ConnectToNSQDs failed, error:", err)
				return err
			}
			l.consumer = append(l.consumer, consumer)
		}
	}
	l.init = true
	return nil
}

func (l *nsqdSvrLogMgr) HandleMessage(message *nsq.Message) error {
	logMsg := &LogMsg{}
	if err := json.Unmarshal(message.Body, logMsg); err != nil {
		fmt.Println("json.Unmarshal Failed:", err)
		return nil
	}
	if l.logLevel > logMsg.Level {
		return nil
	}
	for _, logger := range l.logger {
		if logger.IsValid() {
			logger.Write(logMsg)
		}
	}
	return nil
}

func (l *nsqdSvrLogMgr) Write(*LogMsg) {
	return
}

func (l *nsqdSvrLogMgr) Deinit() {
	if l.init == false {
		return
	}
	for _, c := range l.consumer {
		c.Stop()
	}
	l.waitGroup.Wait()
	for _, logger := range l.logger {
		logger.Deinit()
	}
	l.init = false
}

func (l *nsqdSvrLogMgr) IsValid() bool {
	return l.init
}
