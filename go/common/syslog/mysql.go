package syslog

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosexy/db"
	_ "github.com/gosexy/db/mysql"
	sync_ "sirendaou.com/duserver/common/sync"
)

type MysqlLogMgr struct {
	init      bool
	waitGroup *sync_.WaitGroup
	logMsgCh  chan *LogMsg
	logLevel  int
	logSql    chan string
	session   db.Database
	lastDate  string
}

var (
	drop_log_table       = `DROP TABLE IF EXISTS %v;`
	create_log_table_sql = `CREATE TABLE IF NOT EXISTS %v (id bigint(20) unsigned NOT NULL AUTO_INCREMENT, filename  varchar(96) NOT NULL DEFAULT '', time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP(), name  varchar(96) NOT NULL DEFAULT '', level varchar(12) NOT NULL DEFAULT 'ALL_LEVEL',content varchar(2048),PRIMARY KEY (id) ) DEFAULT CHARSET=utf8;`
)

func (l *MysqlLogMgr) getLogTableName() string {
	curDate := time.Now().Format("2006_01_02")
	tableName := "t_syslog_" + curDate
	if l.lastDate != curDate {
		db := l.session.Driver().(*sql.DB)
		oldTableName := "t_syslog_" + time.Now().AddDate(0, 1, 0).Format("2006_01_02")
		if _, err := db.Exec(fmt.Sprintf(drop_log_table, oldTableName)); err != nil {
			fmt.Println(err)
		}
		/*
			if _, err := db.Exec(fmt.Sprintf(drop_log_table, tableName)); err != nil {
				fmt.Println(err)
			}
		*/
		if _, err := db.Exec(fmt.Sprintf(create_log_table_sql, tableName)); err != nil {
			fmt.Println(err)
		}
		l.lastDate = curDate
	}
	return tableName
}

func init() {
	RegisterLogger(MYSQL, &MysqlLogMgr{})
}
func (l *MysqlLogMgr) Init(config *Config) error {
	if config.MysqlLog == false {
		return errors.New("mysql log server not configure!!!")
	}

	settings := db.DataSource{
		Host:     config.Host,
		Port:     config.Port,
		Database: config.Database,
		User:     config.User,
		Password: config.Password,
		Charset:  config.Charset,
	}

	session, err := db.Open("mysql", settings)
	if err != nil {
		return err
	}
	db := session.Driver().(*sql.DB)
	if err := db.Ping(); err != nil {
		fmt.Println("db.Ping failed:", err)
		return err
	}

	l.init = true
	l.waitGroup = sync_.NewWaitGroup()
	l.logMsgCh = make(chan *LogMsg, getChildLogScaleSize())
	l.logSql = make(chan string, 10)
	l.logLevel = config.MysqlLogLevel
	l.session = session

	for i := 0; i < 10; i++ {
		go l.run()
	}
	go l.writeDB()
	return nil
}

func (l *MysqlLogMgr) run() {
	l.waitGroup.AddOne()
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()
	tick := time.Tick(time.Minute * 5)

	//fmt.Println("db log server start...")
	values := ""
	max_size := 15 * 1024 * 1024
	for {
		select {
		case <-exitNotify:
			return
		case msg := <-l.logMsgCh:
			content := strings.Replace(msg.Content, `'`, `\'`, -1)
			values += fmt.Sprintf("(0,'%v','%v','%v','%v','%v'),", msg.Filename, msg.Time, msg.Name, levelString(msg.Level), content)
			if len(values) > max_size {
				l.logSql <- values
				values = ""
			}
		case <-tick:
			if len(values) > 0 {
				l.logSql <- values
				values = ""
			}
		}
	}
}

func (l *MysqlLogMgr) writeDB() {
	l.waitGroup.AddOne()
	defer l.waitGroup.Done()
	exitNotify := l.waitGroup.ExitNotify()
	db := l.session.Driver().(*sql.DB)
	for {
		select {
		case <-exitNotify:
			return
		case values := <-l.logSql:
			values = strings.TrimRight(values, ",")
			insertSql := fmt.Sprintf("insert delayed %v values", l.getLogTableName()) + values
			if _, err := db.Exec(insertSql); err != nil {
				fmt.Println(insertSql, err)
			}
		}
	}
}

func (l *MysqlLogMgr) Write(logMsg *LogMsg) {
	if l.logLevel <= logMsg.Level {
		l.logMsgCh <- logMsg
	}
}

func (l *MysqlLogMgr) Deinit() {
	if l.init == false {
		return
	}
	l.waitGroup.Wait()
	l.init = false
	l.session.Close()
}

func (l *MysqlLogMgr) IsValid() bool {
	return l.init
}
