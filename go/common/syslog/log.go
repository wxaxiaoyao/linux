package syslog

import (
	"fmt"
	"runtime"
	"time"

	sync_ "sirendaou.com/duserver/common/sync"
)

// log level
const (
	ALL_LEVEL = iota
	DEBUG
	DEBUG1
	DEBUG2
	DEBUG3
	DEBUG4
	DEBUG5
	DEBUG6
	DEBUG7
	DEBUG8
	DEBUG9
	DEBUG10
	DEBUG11
	DEBUG12
	DEBUG13
	DEBUG14
	DEBUG15
	DEBUG16
	INFO  = 32
	WARN  = 64
	ERROR = 128
	FATAL = 256
	OFF   = 512
)

// log driver name
const (
	CONSOLE     = "console"
	FILE        = "file"
	NSQD_CLIENT = "nsqd_client"
	NSQD_SERVER = "nsqd_server"
	MYSQL       = "mysql"
	MONGO       = "mongo"

	DATEFORMAT     = "2006-01-02"
	LOG_SCALE_SIZE = 100000
)

func getCurrentDate() string {
	return time.Now().Format(DATEFORMAT)
}

type LogMsg struct {
	Name     string
	Time     string
	Filename string
	Level    int
	Content  string
	//Comment  string // 备注
}

type logMgr struct {
	init      bool
	logLevel  int
	modelName string
	logMsgCh  chan *LogMsg
	waitGroup *sync_.WaitGroup
	// 日志输出
	logger map[string]Logger
}

type Logger interface {
	Init(config *Config) error
	Write(msg *LogMsg)
	IsValid() bool
	Deinit()
}
type Config struct {
	ModelName string
	LogLevel  int
	// 控制台日志
	ConsoleLog      bool
	ConsoleLogLevel int
	// 文件日志
	FileLog      bool
	FileName     string
	Directory    string
	FileLogLevel int
	// nsqd客户端日志
	NsqdLog      bool
	NsqdAddrs    string
	NsqdTopic    string
	NsqdChannel  string
	NsqdLogLevel int
	// nsqd服务端日志
	NsqdSvrLog    bool
	NsqdSvrConfig *Config
	// DB mysql日志
	MysqlLog      bool
	MysqlLogLevel int
	Driver        string
	Host          string
	Port          int
	Database      string
	User          string
	Password      string
	Charset       string
	// DB mongo日志
	MongoLog      bool
	MongoLogLevel int
	MongoAddr     string
}

var (
	g_sysLogger = &logMgr{
		init:      false,
		logMsgCh:  make(chan *LogMsg, LOG_SCALE_SIZE),
		waitGroup: sync_.NewWaitGroup(),
		logger:    make(map[string]Logger),
	}
	g_defaultConfig = &Config{
		ModelName: "UnKnow",
		LogLevel:  ALL_LEVEL,
		//console log
		ConsoleLog:      false,
		ConsoleLogLevel: ALL_LEVEL,
		//file log
		FileLog:      false,
		FileName:     "syslog.log",
		Directory:    ".",
		FileLogLevel: ALL_LEVEL,
		//nsqd_client log
		NsqdLog:      false,
		NsqdAddrs:    "127.0.0.1:4150",
		NsqdTopic:    "sysLogTopic",
		NsqdChannel:  "syslog_channel",
		NsqdLogLevel: ALL_LEVEL,
		//nsqd_server log
		NsqdSvrLog: false,
		NsqdSvrConfig: &Config{
			ModelName: "nsqd_log_server",
			LogLevel:  ALL_LEVEL,
			//console log
			ConsoleLog:      false,
			ConsoleLogLevel: ALL_LEVEL,
			//file log
			FileLog:      false,
			FileName:     "nsqd_svr_syslog.log",
			Directory:    ".",
			FileLogLevel: ALL_LEVEL,
			//nsqd_client log
			NsqdLog:      false,
			NsqdAddrs:    "127.0.0.1:4150",
			NsqdTopic:    "sysLogTopic",
			NsqdChannel:  "syslog_channel",
			NsqdLogLevel: ALL_LEVEL,
			//nsqd_server log
			NsqdSvrLog:    false,
			NsqdSvrConfig: nil,
			//DB mysql log
			MysqlLog:      false,
			MysqlLogLevel: ALL_LEVEL,
			Driver:        "mysql",
			Host:          "127.0.0.1",
			Port:          3306,
			Database:      "LogDB",
			User:          "root",
			Password:      "root",
			Charset:       "utf8",
			// DB mongo log
			MongoLog:      false,
			MongoLogLevel: ALL_LEVEL,
			MongoAddr:     "127.0.0.1:27017",
		},
		//DB mysql log
		MysqlLog:      false,
		MysqlLogLevel: ALL_LEVEL,
		Driver:        "mysql",
		Host:          "127.0.0.1",
		Port:          3306,
		Database:      "LogDB",
		User:          "root",
		Password:      "root",
		Charset:       "utf8",
		// DB mongo log
		MongoLog:      false,
		MongoLogLevel: ALL_LEVEL,
		MongoAddr:     "127.0.0.1:27017",
	}
)

func getChildLogScaleSize() int {
	return LOG_SCALE_SIZE / 10
}
func getLogScaleSize() int {
	return LOG_SCALE_SIZE
}

func RegisterLogger(logName string, logger Logger) {
	g_sysLogger.logger[logName] = logger
}

func NewConfig() *Config {
	cfg := &Config{}
	nsqdCfg := &Config{}
	defaultNsqdCfg := g_defaultConfig.NsqdSvrConfig
	*cfg = *g_defaultConfig
	*nsqdCfg = *defaultNsqdCfg
	cfg.NsqdSvrConfig = nsqdCfg
	return cfg
}

func SysLogInit(config *Config) {
	// 一个模块只创建一个实例
	if g_sysLogger.init {
		return
	}
	if config == nil {
		config = g_defaultConfig
	}
	g_sysLogger.init = true
	g_sysLogger.logLevel = config.LogLevel
	g_sysLogger.modelName = config.ModelName

	for logName, logger := range g_sysLogger.logger {
		//fmt.Printf("init log driver...[%s] [%s]\n", logName, config.ModelName)
		if err := logger.Init(config); err != nil {
			fmt.Printf("init log driver failed!!![%s] [%s]; error:%s\n", logName, config.ModelName, err.Error())
			continue
		}
		//fmt.Printf("init log driver success...[%s] [%s]\n", logName, config.ModelName)
	}

	go func() {
		//fmt.Println("route log routine start...")
		g_sysLogger.waitGroup.Add(1)
		defer g_sysLogger.waitGroup.Done()
		exitNotify := g_sysLogger.waitGroup.ExitNotify()
		for {
			select {
			case <-exitNotify:
				fmt.Println("route log routine stop!!!")
				return
			case msg := <-g_sysLogger.logMsgCh:
				for _, logger := range g_sysLogger.logger {
					if logger.IsValid() {
						logger.Write(msg)
					}
				}
			}
		}
	}()
	//fmt.Println("log server start...")
	return
}

func SysLogDeinit() {
	g_sysLogger.waitGroup.Wait()
	g_sysLogger.init = false
	for _, logger := range g_sysLogger.logger {
		logger.Deinit()
	}
	//fmt.Println("log server stop!!!")
}

func NewLogMsg(level int, content string) *LogMsg {
	_, file, line, _ := runtime.Caller(2)
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	logMsg := &LogMsg{
		Name:     g_sysLogger.modelName,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Filename: fmt.Sprint(short, ":", line),
		Level:    level,
		Content:  content,
	}
	return logMsg
}

func (logMsg *LogMsg) Format() string {
	return fmt.Sprintln(logMsg.Time, " ", logMsg.Filename, " ", logMsg.Name, " ", levelString(logMsg.Level), " ", logMsg.Content)
}

func levelString(level int) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case DEBUG1:
		return "DEBUG1"
	case DEBUG2:
		return "DEBUG2"
	case DEBUG3:
		return "DEBUG3"
	case DEBUG4:
		return "DEBUG4"
	case DEBUG5:
		return "DEBUG5"
	case DEBUG6:
		return "DEBUG6"
	case DEBUG7:
		return "DEBUG7"
	case DEBUG8:
		return "DEBUG8"
	case DEBUG9:
		return "DEBUG9"
	case DEBUG10:
		return "DEBUG10"
	case DEBUG11:
		return "DEBUG11"
	case DEBUG12:
		return "DEBUG12"
	case DEBUG13:
		return "DEBUG13"
	case DEBUG14:
		return "DEBUG14"
	case DEBUG15:
		return "DEBUG15"
	case DEBUG16:
		return "DEBUG16"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case ALL_LEVEL:
		return "ALL_LEVEL"
	case OFF:
		return "OFF"
	}
	return "UNKNOW"
}

func sysLog(msg *LogMsg) {
	// 日志未开启
	if g_sysLogger.init == false {
		return
	}
	g_sysLogger.logMsgCh <- msg
}

func Debug(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG {
		sysLog(logMsg)
	}
}

func Debug1(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG1, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG1 {
		sysLog(logMsg)
	}
}
func Debug2(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG2, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG2 {
		sysLog(logMsg)
	}
}
func Debug3(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG3, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG3 {
		sysLog(logMsg)
	}
}
func Debug4(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG4, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG4 {
		sysLog(logMsg)
	}
}
func Debug5(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG5, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG5 {
		sysLog(logMsg)
	}
}
func Debug6(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG6, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG6 {
		sysLog(logMsg)
	}
}
func Debug7(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG7, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG7 {
		sysLog(logMsg)
	}
}
func Debug8(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG8, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG8 {
		sysLog(logMsg)
	}
}
func Debug9(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG9, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG9 {
		sysLog(logMsg)
	}
}
func Debug10(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG10, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG10 {
		sysLog(logMsg)
	}
}
func Debug11(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG11, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG11 {
		sysLog(logMsg)
	}
}
func Debug12(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG12, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG12 {
		sysLog(logMsg)
	}
}
func Debug13(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG13, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG13 {
		sysLog(logMsg)
	}
}
func Debug14(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG14, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG14 {
		sysLog(logMsg)
	}
}
func Debug15(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG15, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG15 {
		sysLog(logMsg)
	}
}
func Debug16(v ...interface{}) {
	logMsg := NewLogMsg(DEBUG16, fmt.Sprint(v))
	if g_sysLogger.logLevel <= DEBUG16 {
		sysLog(logMsg)
	}
}
func Info(v ...interface{}) {
	logMsg := NewLogMsg(INFO, fmt.Sprint(v))
	if g_sysLogger.logLevel <= INFO {
		sysLog(logMsg)
	}
}

func Warn(v ...interface{}) {
	logMsg := NewLogMsg(WARN, fmt.Sprint(v))
	if g_sysLogger.logLevel <= WARN {
		sysLog(logMsg)
	}
}

func Error(v ...interface{}) {
	logMsg := NewLogMsg(ERROR, fmt.Sprint(v))
	if g_sysLogger.logLevel <= ERROR {
		sysLog(logMsg)
	}
}

func Fatal(v ...interface{}) {
	logMsg := NewLogMsg(FATAL, fmt.Sprint(v))
	if g_sysLogger.logLevel <= FATAL {
		sysLog(logMsg)
	}
}
