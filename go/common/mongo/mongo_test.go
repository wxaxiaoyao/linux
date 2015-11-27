package mongo

import (
	"fmt"
	"log"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestMongoDB(t *testing.T) {
	Init("127.0.0.1:27017", 1)

	session := Get()
	defer Put(session)

	c := session.DB("testDB").C("testTable")

	index := mgo.Index{
		Key: []string{"field3:1"},
	}

	if err := c.EnsureIndex(index); err != nil {
		println(err.Error())
		return
	}

	cnt := bson.M{
		"field1": 1,
		"field2": "2",
		"field4": 4,
	}

	if err := c.Insert(cnt); err != nil {
		println(err.Error())
		return
	}

	//MongoClose()
}

type MongoInsert struct {
	Name string
	Age  int
}

type MongoDel struct {
	Name string
}
type MongoModify struct {
	Name string
	Sex  string
}

type MongoAdd struct {
	Name string
	Age  int
	Sex  string
}

func TestStrcut(t *testing.T) {
	Init("127.0.0.1:27017", 1)

	session := Get()
	defer Put(session)

	c := session.DB("testDB").C("testTable1")

	mi := &MongoInsert{"test", 24}

	if err := c.Insert(mi); err != nil {
		fmt.Println(err)
		return
	}

	selector := bson.M{}
	iter := c.Find(selector).Iter()
	m1 := &MongoInsert{}
	if iter.Next(m1) {
		fmt.Println("============1", m1)
	}
	iter = c.Find(selector).Iter()
	m2 := &MongoDel{}
	if iter.Next(m2) {
		fmt.Println("============2", m2)
	}
	iter = c.Find(selector).Iter()
	m3 := &MongoAdd{}
	if iter.Next(m3) {
		fmt.Println("============3", m3)
	}
}

type LogMsg struct {
	Name     string
	Time     string
	Filename string
	Level    int
	Content  string
	//Comment  string // 备注
}

func TestInsert(t *testing.T) {
	Init("127.0.0.1:27017", 1)

	session := Get()
	defer Put(session)
	content := "test"
	for i := 0; i < 220; i++ {
		content += "test"
	}
	msgs := []interface{}{}
	msg := &LogMsg{
		Name:     "test",
		Time:     "test",
		Filename: "test",
		Level:    0,
		Content:  content,
	}
	for i := 0; i < 15000; i++ {
		msgs = append(msgs, msg)
	}

	c := session.DB("testDB").C("syslog")
	log.Println("start")
	for i := 0; i < 800; i++ {
		log.Println("start", i)
		if err := c.Insert(msgs...); err != nil {
			log.Println(err)
			return
		}
		log.Println("end", i)
	}
	log.Println("end")
	time.Sleep(time.Second)
	Deinit()
}
