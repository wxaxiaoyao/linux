package syslog

import (
	"testing"
	"time"
)

func TestNsqClient(t *testing.T) {
	config := NewConfig()
	config.ConsoleLog = false
	config.FileLog = false
	config.MysqlLog = false
	config.MongoLog = false
	config.NsqdLog = true
	config.NsqdSvrLog = true
	config.NsqdTopic = "testlog"
	SysLogInit(config)

	for i := 0; i <= 30; i++ {
		Debug("test log performance")
	}
	time.Sleep(time.Second * 10)
	SysLogDeinit()
	time.Sleep(time.Second * 10)
}
