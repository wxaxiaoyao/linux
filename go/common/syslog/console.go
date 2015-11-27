package syslog

import (
	"errors"
	"fmt"

	sync_ "sirendaou.com/duserver/common/sync"
)

type consoleLogMgr struct {
	init      bool
	waitGroup *sync_.WaitGroup
	logMsgCh  chan *LogMsg
	logLevel  int
}

func init() {
	RegisterLogger(CONSOLE, &consoleLogMgr{})
}
func (l *consoleLogMgr) Init(config *Config) error {
	if config.ConsoleLog == false {
		return errors.New("console log server not configure!!!")
	}

	l.init = true
	l.waitGroup = sync_.NewWaitGroup()
	l.logMsgCh = make(chan *LogMsg, getChildLogScaleSize())
	l.logLevel = config.ConsoleLogLevel

	go l.run()
	return nil
}

func (l *consoleLogMgr) run() {
	l.waitGroup.AddOne()
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()

	//fmt.Println("console log server start...")
	for {
		select {
		case <-exitNotify:
			//fmt.Println("console log server stop!!!")
			return
		case msg := <-l.logMsgCh:
			fmt.Print(msg.Format())
		}
	}
}

func (l *consoleLogMgr) Write(logMsg *LogMsg) {
	if l.logLevel <= logMsg.Level {
		l.logMsgCh <- logMsg
	}
}

func (l *consoleLogMgr) Deinit() {
	if l.init == false {
		return
	}
	l.waitGroup.Wait()
	l.init = false
}

func (l *consoleLogMgr) IsValid() bool {
	return l.init
}
