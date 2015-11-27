package mysql

import (
	"database/sql"

	"github.com/gosexy/db"
	_ "github.com/gosexy/db/mysql"

	"sirendaou.com/duserver/common/errors"
	//"sirendaou.com/duserver/common/syslog"
)

const (
	mysql_conn_max_num = 10
	mysql_conn_min_num = 1
)

var (
	ErrNoRows = errors.New("mysql select have no record!!!")
)

type MysqlManager struct {
	dbCh  chan *sql.DB
	count int
}

var g_mysql *MysqlManager = nil

func Init(host, dbname, user, passwd string, count int) {
	if count > mysql_conn_max_num {
		count = mysql_conn_max_num
	} else if count < mysql_conn_min_num {
		count = mysql_conn_min_num
	}

	var settings = db.DataSource{
		Host:     host,
		Database: dbname,
		User:     user,
		Password: passwd,
	}

	dbCh := make(chan *sql.DB, count)

	for i := 0; i < count; i++ {
		sess, err := db.Open("mysql", settings)
		if err != nil {
			panic(err)
		}

		drv := sess.Driver().(*sql.DB)
		if err := drv.Ping(); err != nil {
			panic(err)
		}

		dbCh <- drv
	}

	g_mysql = &MysqlManager{dbCh, count}

	return
}

func Get() *sql.DB {
	sqlDB := <-g_mysql.dbCh
	return sqlDB
}

func Put(sqlDB *sql.DB) {
	g_mysql.dbCh <- sqlDB
}

func Deinit() {
	for i := 0; i < g_mysql.count; i++ {
		sqlDB := <-g_mysql.dbCh
		sqlDB.Close()
	}
}

func Exec(sqlStr string, args ...interface{}) error {
	sqlDB := Get()
	defer Put(sqlDB)

	if _, err := sqlDB.Exec(sqlStr, args...); err != nil {
		return errors.As(err, sqlStr, args)
	}
	return nil
}

func ExecRet(sqlStr string, args ...interface{}) (uint64, error) {
	sqlDB := Get()
	defer Put(sqlDB)

	ret, err := sqlDB.Exec(sqlStr, args...)
	if err != nil {
		return 0, errors.As(err, sqlStr, args)
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return 0, errors.As(err, sqlStr, args)
	}

	return uint64(id), nil
}

func Query(sqlStr string, args ...interface{}) (*sql.Rows, error) {
	sqlDB := Get()
	defer Put(sqlDB)

	rows, err := sqlDB.Query(sqlStr, args...)
	if err != nil {
		return nil, errors.As(err, sqlStr, args)
	}

	return rows, nil
}
