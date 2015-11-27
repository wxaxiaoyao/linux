package syslog

import (
	"testing"
	"time"
)

func TestSysLog(t *testing.T) {
	config := NewConfig()
	config.NsqdTopic = "testTopic"
	config.NsqdLog = true
	config.NsqdSvrLog = true
	config.NsqdSvrConfig.ConsoleLog = true

	SysLogInit(config)
	//fmt.Println(config)
	//fmt.Println(config.NsqdSvrConfig)
	Debug("hello world")
	Info("hello world")

	time.Sleep(time.Second * 5)
	SysLogDeinit()

}
