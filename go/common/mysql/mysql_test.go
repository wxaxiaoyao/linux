package mysql

import (
	"fmt"
	"log"
	"testing"
)

func TestMysqlInsert(t *testing.T) {
	Init("127.0.0.1", "logDB", "root", "Youkang@0814", 1)

	db := Get()

	content := "testa"
	for i := 0; i < 200; i++ {
		content += "yrday"
	}
	sql := "insert delayed t_syslog values"

	for i := 0; i < 15000; i++ {
		sql += fmt.Sprintf("(0,'%v','%v','%v','%v','%v'),", "test", "test", "test", "test", content)
	}
	sql += fmt.Sprintf("(0,'%v','%v','%v','%v','%v')", "test", "test", "test", "test", "test")

	log.Println("==========start=========")
	for i := 0; i < 80; i++ {
		log.Println("start:", i)
		//if _, err := db.Exec("lock tables t_syslog write"); err != nil {
		/*
			tx, err := db.Begin()
			if err != nil {
				fmt.Println(err)
				return
			}
		*/
		for j := 0; j < 10; j++ {
			if _, err := db.Exec(sql); err != nil {
				log.Println(err)
				return
			}
		}
		//if _, err := db.Exec("unlock tables"); err != nil {
		/*
			if err := tx.Commit(); err != nil {
				fmt.Println(err)
				return
			}
		*/
		log.Println("end:", i)
	}
	log.Println("==========end=========")
	Put(db)
	Deinit()
	// 1 次插入 2w条记录
}
