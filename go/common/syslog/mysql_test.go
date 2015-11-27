package syslog

import (
	"testing"
	"time"
)

func TestDBLog(t *testing.T) {
	l := MysqlLogMgr{}
	config := NewConfig()
	config.MysqlLog = true
	config.Database = "logDB"
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

func TestCreateLogTable(t *testing.T) {
	l := MysqlLogMgr{}
	config := NewConfig()
	config.MysqlLog = true
	config.Database = "logDB"
	l.Init(config)
	l.Deinit()
}
