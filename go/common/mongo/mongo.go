package mongo

import (
	"gopkg.in/mgo.v2"
)

type MongoManager struct {
	sessCh chan *mgo.Session
	count  int
}

const (
	mongo_conn_max_num = 10
	mongo_conn_min_num = 1
)

var g_mongo *MongoManager = nil

func Init(mongodbAddr string, count int) {
	if count > mongo_conn_max_num {
		count = mongo_conn_max_num
	} else if count < mongo_conn_min_num {
		count = mongo_conn_min_num
	}

	sessCh := make(chan *mgo.Session, count)

	for i := 0; i < count; i++ {
		sess, err := mgo.Dial(mongodbAddr)
		if err != nil {
			panic(err)
		}

		//Optional. Switch the session to a monotonic behavior.
		sess.SetMode(mgo.Monotonic, true)
		sessCh <- sess
	}

	g_mongo = &MongoManager{
		sessCh: sessCh,
		count:  count,
	}

	return
}

func Get() *mgo.Session {
	session := <-g_mongo.sessCh
	return session
}

func Put(session *mgo.Session) {
	g_mongo.sessCh <- session
}

func Deinit() {
	for i := 0; i < g_mongo.count; i++ {
		sess := <-g_mongo.sessCh
		sess.Close()
	}
	close(g_mongo.sessCh)
}

func Collection(db, table string) *mgo.Collection {
	session := Get()
	defer Put(session)
	return session.DB(db).C(table)
}
