package user

import (
	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/syslog"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Around System
const (
	LOCAL_MAX_PAGE = 30
	MAX_ROW        = 28 //每页数据最大值
)

type LocationInfo struct {
	Uid uint64 `bson:"uid"`
	//	AppKey      string    `bson:"appkey"`
	//	DeveloperID int       `bson:"developerID"`
	Loc      []float64 `bson:"loc"`
	SendTime uint32    `bson:"sendtime"`
}

type LocResult struct {
	Uid  uint64  `json:"uid"`
	Xpos float64 `json:"xpos"`
	Ypos float64 `json:"ypos"`
}

func (location *LocationInfo) NewLocation() error {
	session := common.MongoGet()
	defer common.MongoPut(session)

	c := session.DB("du").C("location")

	index := mgo.Index{
		Key:        []string{"$2d:loc"},
		Bits:       26,
		Background: true,
	}
	if err := c.EnsureIndex(index); err != nil {
		return errors.As(err, index)
	}

	newLoc := bson.M{
		"uid":      location.Uid,
		"loc":      location.Loc,
		"sendtime": location.SendTime,
	}

	if err := c.Insert(newLoc); err != nil {
		return errors.As(err, newLoc)
	}

	return nil
}

func (location *LocationInfo) GetLocationInfo() error {
	session := common.MongoGet()
	defer common.MongoPut(session)

	c := session.DB("du").C("location")

	selector := bson.M{"uid": location.Uid}

	iter := c.Find(selector).Iter()

	for iter.Next(&location) {
		syslog.Info("Uid: ", location.Uid, " AppKey: ", " Loc: ", location.Loc, " send time: ", location.SendTime)
	}

	if err := iter.Close(); err != nil {
		return errors.As(err, *location)
	}

	return nil
}

func (location *LocationInfo) SaveLocation() error {
	session := common.MongoGet()
	defer common.MongoPut(session)

	c := session.DB("du").C("location")

	selector := bson.M{"uid": location.Uid}
	newLoc := bson.M{"uid": location.Uid, "loc": location.Loc, "sendtime": location.SendTime}
	update := bson.M{"$set": newLoc}
	syslog.Debug("selector: ", selector, " update: ", update)
	if err := c.Update(selector, update); err != nil {
		return errors.As(err, *location)
	}

	return nil
}

func (location *LocationInfo) GetLocation(level uint16, hour uint32, page uint16) []LocResult {
	if level != 1 {
		syslog.Error("level ", level, " not completed.")
		return nil
	}

	session := common.MongoGet()
	defer common.MongoPut(session)

	c := session.DB("du").C("location")

	selector := bson.M{"loc": bson.M{"$near": location.Loc}}
	syslog.Debug("selector: ", selector)
	iter := c.Find(selector).Limit(MAX_ROW*LOCAL_MAX_PAGE + 1).Iter()
	result := LocationInfo{}

	retLoc := make([]LocResult, MAX_ROW*30)
	line := 0

	retLoc[line].Uid = location.Uid
	retLoc[line].Xpos = float64(int64(location.Loc[0]*1000000)) / 1000000
	retLoc[line].Ypos = float64(int64(location.Loc[1]*1000000)) / 1000000
	syslog.Debug(retLoc[line].Xpos, retLoc[line].Ypos)
	line++

	for iter.Next(&result) && line < MAX_ROW*30 {
		if result.Uid == location.Uid {
			continue
		}

		syslog.Debug("Uid: ", result.Uid, " Loc: ", result.Loc, " send time: ", result.SendTime, "line:", line)

		retLoc[line].Uid = result.Uid
		retLoc[line].Xpos = float64(int64(result.Loc[0]*1000000)) / 1000000
		retLoc[line].Ypos = float64(int64(result.Loc[1]*1000000)) / 1000000
		//retlocl[line].Xpos = result.Loc[0]
		//retlocl[line].Ypos = result.Loc[1]

		line++
	}

	return retLoc[:line]
}
