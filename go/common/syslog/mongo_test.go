package syslog

import (
	//"log"
	"testing"
	"time"
)

func TestMongoLog(t *testing.T) {
	l := MongoLogMgr{}
	config := NewConfig()
	config.MongoLog = true
	l.Init(config)

	msg := &LogMsg{}
	msg.Content = "hello world"
	msg.Filename = "db_test:17"
	msg.Level = 1
	msg.Name = "syslog"
	msg.Time = time.Now().Format(DATEFORMAT)

	l.Write(msg)

	time.Sleep(time.Second * 3)

	l.Deinit()
}
