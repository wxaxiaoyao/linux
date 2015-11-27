package syslog

import (
	"errors"
	"fmt"
	"os"
	"time"

	sync_ "sirendaou.com/duserver/common/sync"
)

type fileLogMgr struct {
	init      bool
	waitGroup *sync_.WaitGroup
	logMsgCh  chan *LogMsg
	logLevel  int
	directory string
	filename  string
	date      time.Time
	file      *os.File
}

func init() {
	RegisterLogger(FILE, &fileLogMgr{})
}

func (l *fileLogMgr) Init(config *Config) error {
	if config.FileLog == false {
		return errors.New("file log server not configure!!!")
	}

	date, err := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	if err != nil {
		fmt.Println("time.Parse failed:", err)
		return err
	}
	dir := config.Directory
	if len(dir) > 0 && dir[len(dir)-1] != '/' {
		dir += "/"
	}
	file, err := os.OpenFile(dir+config.FileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("os.OpenFile failed", err)
		return err
	}

	l.waitGroup = sync_.NewWaitGroup()
	l.logMsgCh = make(chan *LogMsg, getChildLogScaleSize())
	l.logLevel = config.FileLogLevel
	l.filename = config.FileName
	l.date = date
	l.directory = dir
	l.file = file
	l.init = true
	go l.run()

	return nil
}

func (l *fileLogMgr) run() {
	l.waitGroup.AddOne()
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()

	//fmt.Println("file log server start...")
	for {
		select {
		case <-exitNotify:
			//fmt.Println("file log server stop!!!")
			return
		case msg := <-l.logMsgCh:
			l.check()
			l.file.WriteString(msg.Format())
		}
	}
}

func (l *fileLogMgr) Write(logMsg *LogMsg) {
	if l.logLevel <= logMsg.Level {
		l.logMsgCh <- logMsg
	}
}

func (l *fileLogMgr) Deinit() {
	if l.init == false {
		return
	}
	l.waitGroup.Wait()
	l.file.Close()
	l.init = false
}

func (l *fileLogMgr) IsValid() bool {
	return l.init
}
func (l *fileLogMgr) check() {
	filename := l.directory + l.filename + "." + l.date.Format(DATEFORMAT)
	if l.isMustRename() && !isExist(filename) {
		if l.file != nil {
			l.file.Close()
		}
		err := os.Rename(l.directory+l.filename, filename)
		if err != nil {
			fmt.Println("os.Rename failed:", err)
			// WARN
		}
		file, err := os.OpenFile(l.directory+l.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("os.OpenFile failed:", err)
			return
		}
		date, err := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if err != nil {
			fmt.Println("time.Parse failed:", err)
			return
		}
		l.file = file
		l.date = date
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func (l *fileLogMgr) isMustRename() bool {
	t, err := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
	if err != nil {
		fmt.Println("time.Parse failed:", err)
	}
	if t.After(l.date) {
		return true
	}
	return false
}
