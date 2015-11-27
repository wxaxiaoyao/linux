package common

import (
	"database/sql"
	"flag"
	"runtime"

	"github.com/rakyll/globalconf"
	"gopkg.in/mgo.v2"

	"sirendaou.com/duserver/common/mongo"
	"sirendaou.com/duserver/common/mysql"
	"sirendaou.com/duserver/common/nsq"
	"sirendaou.com/duserver/common/redis"
	"sirendaou.com/duserver/common/syslog"
)

//========================conf common=====================
var (
	// hardware profile
	g_conf_file   = flag.String("conf_file", "", "the program config file path")
	g_cpu_num     = flag.Int("cpu_num", 4, "the num of cpu")
	g_module_name = flag.String("module_name", "unknow_module_name", "the program name")
	// log enable profile
	g_log_console_enable    = flag.Bool("log_console_enable", true, "the log output to console")
	g_log_file_enable       = flag.Bool("log_file_enable", false, "the log output to file")
	g_log_mysql_enable      = flag.Bool("log_mysql_enable", false, "the log output to mysql database")
	g_log_mongo_enable      = flag.Bool("log_mongo_enable", false, "the log output to mongo database")
	g_log_nsq_client_enable = flag.Bool("log_nsq_client_enable", false, "the log output to nsq_server")
	g_log_nsq_server_enable = flag.Bool("log_nsq_server_enable", false, "recv from  nsq_client log")
	// log level profile
	g_log_level            = flag.Int("log_level", syslog.ALL_LEVEL, "the level of log")
	g_log_console_level    = flag.Int("log_console_level", syslog.ALL_LEVEL, "the level of console log")
	g_log_file_level       = flag.Int("log_file_level", syslog.ALL_LEVEL, "the level of file log")
	g_log_mysql_level      = flag.Int("log_mysql_level", syslog.ALL_LEVEL, "the level of mysql database log")
	g_log_mongo_level      = flag.Int("log_mongo_level", syslog.ALL_LEVEL, "the level of mongo database log")
	g_log_nsq_client_level = flag.Int("log_nsq_client_level", syslog.ALL_LEVEL, "the level of nsq_client log")
	g_log_nsq_server_level = flag.Int("log_nsq_server_level", syslog.ALL_LEVEL, "the level of nsq_server log")
	// log file profile
	g_log_path = flag.String("log_path", "../log", "the log file path")
	g_log_file = flag.String("log_file", "server.log", "the log file path")
	// log db profile
	g_mysql_log_db = flag.String("mysql_log_db", "logDB", "mysql log db name")
	// log nsq addrs
	g_log_nsqd_addrs = flag.String("log_nsq_addr", "127.0.0.1:4150", "log nsq Server address (transient)")
	// mysql profile
	g_mysql_host = flag.String("mysql_host", "127.0.0.1", "mysql host")
	g_mysql_db   = flag.String("mysql_db", "dudb", "mysql db name")
	g_mysql_user = flag.String("mysql_user", "root", "mysql user")
	g_mysql_pwd  = flag.String("mysql_pwd", "Youkang@0814", "mysql passwd")

	// server addr profile
	g_RedisAddr   = flag.String("redis_addr", "127.0.0.1:6379", "redis mq server addr")
	g_MongodbAddr = flag.String("mongodb_addr", "127.0.0.1:27017", "mongodb  server addr")
	g_nsqd_addrs  = flag.String("nsq_addr", "127.0.0.1:4150", "nsq Server address (transient)")
)

func ConfInit() {
	flag.Parse()
	if *g_conf_file == "" {
		//panic("please set conf file ")
		println("have no config file")
		return
	}
	conf, err := globalconf.NewWithOptions(&globalconf.Options{
		Filename: *g_conf_file,
	})

	if err != nil {
		panic(err)
	}
	conf.ParseAll()
}

func syslogInit() {
	config := syslog.NewConfig()
	config.ModelName = *g_module_name
	config.LogLevel = *g_log_level
	//config.LogLevel = syslog.ALL_LEVEL

	config.ConsoleLog = *g_log_console_enable
	config.ConsoleLogLevel = *g_log_console_level

	config.FileLog = *g_log_file_enable
	config.FileName = *g_log_file
	config.Directory = *g_log_path
	config.FileLogLevel = *g_log_file_level

	config.MysqlLog = *g_log_mysql_enable
	config.MysqlLogLevel = *g_log_mysql_level
	config.Host = *g_mysql_host
	config.Database = *g_mysql_log_db
	config.User = *g_mysql_user
	config.Password = *g_mysql_pwd

	config.MongoLog = *g_log_mongo_enable
	config.MongoLogLevel = *g_log_mongo_level
	config.MongoAddr = *g_MongodbAddr

	config.NsqdLog = *g_log_nsq_client_enable
	config.NsqdAddrs = *g_log_nsqd_addrs
	config.NsqdLogLevel = *g_log_nsq_client_level
	if *g_log_nsq_server_enable == true {
		config.FileLog = false
		config.ConsoleLog = false
		config.MysqlLog = false
		config.MongoLog = false
		config.NsqdLog = false
	}

	config.NsqdSvrLog = *g_log_nsq_server_enable
	config.NsqdSvrConfig.FileLog = *g_log_file_enable
	config.NsqdSvrConfig.FileName = *g_log_file
	config.NsqdSvrConfig.Directory = *g_log_path
	config.NsqdSvrConfig.FileLogLevel = *g_log_file_level

	config.NsqdSvrConfig.NsqdLog = false // 必须false 避免死循环日志 *g_log_nsq_client_enable
	config.NsqdSvrConfig.NsqdAddrs = *g_log_nsqd_addrs
	config.NsqdSvrConfig.NsqdLogLevel = *g_log_nsq_client_level

	config.NsqdSvrConfig.MysqlLog = *g_log_mysql_enable
	config.NsqdSvrConfig.Database = *g_mysql_log_db
	config.NsqdSvrConfig.MysqlLogLevel = *g_log_mysql_level
	config.NsqdSvrConfig.Host = *g_mysql_host
	config.NsqdSvrConfig.User = *g_mysql_user
	config.NsqdSvrConfig.Password = *g_mysql_pwd

	config.NsqdSvrConfig.MongoLog = *g_log_mongo_enable
	config.NsqdSvrConfig.MongoLogLevel = *g_log_mongo_level
	config.NsqdSvrConfig.MongoAddr = *g_MongodbAddr

	config.NsqdSvrConfig.ConsoleLog = *g_log_console_enable
	config.NsqdSvrConfig.ConsoleLogLevel = *g_log_console_level
	syslog.SysLogInit(config)
	syslog.Info("syslog init success!")
}

func init() {
	ConfInit()
	runtime.GOMAXPROCS(*g_cpu_num)
	syslogInit()
	MysqlInit(*g_mysql_host, *g_mysql_db, *g_mysql_user, *g_mysql_pwd)
	RedisInit(*g_RedisAddr)
	MongoInit(*g_MongodbAddr)
	NsqInit(*g_nsqd_addrs)
}

//========================mongo common====================
func MongoInit(mongodbAddr string) {
	mongo.Init(mongodbAddr, 2)
}
func MongoGet() *mgo.Session {
	return mongo.Get()
}
func MongoPut(sess *mgo.Session) {
	mongo.Put(sess)
}
func MongoCollection(db, table string) *mgo.Collection {
	return mongo.Collection(db, table)
}
func MongoDeinit() {
	mongo.Deinit()
}

//=========================mysql common========================
func MysqlInit(host, dbname, user, passwd string) {
	mysql.Init(host, dbname, user, passwd, 2)
}
func MysqlGet() *sql.DB {
	return mysql.Get()
}
func MysqlPut(sqlDB *sql.DB) {
	mysql.Put(sqlDB)
}
func MysqlExec(sqlStr string, args ...interface{}) error {
	return mysql.Exec(sqlStr, args...)
}
func MysqlExecRet(sqlStr string, args ...interface{}) (uint64, error) {
	return mysql.ExecRet(sqlStr, args...)
}
func MysqlQuery(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	return mysql.Query(sqlStr, args...)
}
func MysqlDeinit() {
	mysql.Deinit()
}

//=======================nsq common===========================
func NsqInit(addrs string) {
	nsq.Init(addrs)
}
func NsqPublish(topic string, body []byte) {
	nsq.Publish(topic, body)
}
func NsqConsumer(topic, channel string, handle nsq.Handler) (*nsq.ConsumerT, error) {
	return nsq.Consumer(topic, channel, handle)
}
func NsqConsumerGO(topic, channel string, goCount uint, handle nsq.Handler) (*nsq.ConsumerT, error) {
	return nsq.ConsumerGO(topic, channel, goCount, handle)
}
func NsqDeinit() {
	nsq.Deinit()
}

//=====================redis common====================
func RedisInit(addrs string) {
	redis.Init(addrs, 10)
}
func RedisDeinit() {
	redis.Deinit()
}
