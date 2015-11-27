package user

import (
	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserDetailInfo struct {
	Uid uint64 `bson:"uid"`
	// 用户终端信息
	Platform string `bson:"platform"`
	DeviceId string `bson:"deviceid"`
	// 用户基本信息
	// sex  string
}

func UserDetailInfoInit() error {
	c := common.MongoCollection("dudb", "user_detail_info")

	index := mgo.Index{
		Key:        []string{"uid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
	}
	if err := c.EnsureIndex(index); err != nil {
		return errors.As(err, index)
	}
	return nil
}

func SaveUserDetailInfo(user *UserDetailInfo) error {
	c := common.MongoCollection("dudb", "user_detail_info")

	if err := c.Insert(user); err != nil {
		return errors.As(err, *user)
	}

	return nil
}

func GetUserDetailInfoByUid(uid uint64) (*UserDetailInfo, error) {
	user := &UserDetailInfo{}
	c := common.MongoCollection("dudb", "user_detail_info")

	if err := c.Find(bson.M{"uid": uid}).One(user); err != nil {
		return nil, errors.As(err, uid)
	}
	return user, nil
}

func DeleteUserDetailInfoByUid(uid uint64) error {
	c := common.MongoCollection("dudb", "user_detail_info")
	if err := c.Remove(bson.M{"uid": uid}); err != nil {
		return errors.As(err, uid)
	}
	return nil
}

func SetUserDetailInfo(sel, set map[string]interface{}) error {
	c := common.MongoCollection("dudb", "user_detail_info")
	if err := c.Update(bson.M(sel), bson.M{"$set": bson.M(set)}); err != nil {
		return errors.As(err, sel, set)
	}
	return nil
}
