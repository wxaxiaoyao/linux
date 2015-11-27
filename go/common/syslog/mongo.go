package syslog

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	sync_ "sirendaou.com/duserver/common/sync"
)

type MongoLogMgr struct {
	init      bool
	waitGroup *sync_.WaitGroup
	logMsgCh  chan *LogMsg
	logLevel  int

	session  *mgo.Session
	lastDate string
}

var (
	DB_NAME          = "logDB"
	LOG_TABLE_PREFIX = "t_syslog_"
)

func (l *MongoLogMgr) getCollection() *mgo.Collection {
	curDate := time.Now().Format("2006_01_02")
	tableName := LOG_TABLE_PREFIX + curDate
	if l.lastDate != curDate {
		oldTableName := LOG_TABLE_PREFIX + time.Now().AddDate(0, 1, 0).Format("2006_01_02")
		if err := l.session.DB(DB_NAME).C(oldTableName).DropCollection(); err != nil {
			//fmt.Println("dropCollection:", err)
		}
		l.lastDate = curDate
	}
	return l.session.DB(DB_NAME).C(tableName)
}

func init() {
	RegisterLogger(MONGO, &MongoLogMgr{})
}
func (l *MongoLogMgr) Init(config *Config) error {
	if config.MongoLog == false {
		return errors.New("mongo log server not configure!!!")
	}

	session, err := mgo.Dial(config.MongoAddr)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)

	l.init = true
	l.waitGroup = sync_.NewWaitGroup()
	l.logMsgCh = make(chan *LogMsg, getChildLogScaleSize())
	l.logLevel = config.MongoLogLevel
	l.session = session

	go l.run()
	return nil
}

func (l *MongoLogMgr) run() {
	l.waitGroup.AddOne()
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()
	msgCacheCount := 20000
	msgSlice := make([]interface{}, msgCacheCount)
	msgCount := 0

	tick := time.Tick(time.Minute * 15)
	//tick := time.Tick(time.Millisecond * 15)
	//fmt.Println("mongo log server start...")
	for {
		select {
		case <-exitNotify:
			//fmt.Println("mongo log server stop!!!")
			if msgCount > 0 {
				if err := l.getCollection().Insert(msgSlice[:msgCount]); err != nil {
					fmt.Println(err)
				}
				msgCount = 0
			}
			return
		case msg := <-l.logMsgCh:
			msgSlice[msgCount] = msg
			msgCount++
			if msgCount == msgCacheCount {
				if err := l.getCollection().Insert(msgSlice[:msgCount]); err != nil {
					fmt.Println(err)
				}
				msgCount = 0
			}
		case <-tick:
			if msgCount > 0 {
				if err := l.getCollection().Insert(msgSlice[:msgCount]); err != nil {
					fmt.Println(err)
				}
				msgCount = 0
			}
		}
	}
}
func (l *MongoLogMgr) Write(logMsg *LogMsg) {
	if l.logLevel <= logMsg.Level {
		l.logMsgCh <- logMsg
	}
}

func (l *MongoLogMgr) Deinit() {
	if l.init == false {
		return
	}
	l.waitGroup.Wait()
	l.init = false
	l.session.Close()
}

func (l *MongoLogMgr) IsValid() bool {
	return l.init
}
