package user

import (
	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserInfo struct {
	Uid          uint64
	Password     string
	PhoneNum     string
	RegisterDate string
	UpdateDate   string
	// 用户终端信息
	Platform string `bson:"platform"`
	DeviceId string `bson:"deviceid"`
	SetupId  string `bson:"setupid"`
	BaseInfo string
	ExInfo   string
	// 用户基本信息
	// sex  string
}

func UserInfoInit() error {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)

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

//User System
func SaveUserInfo(user *UserInfo) error {
	user_mutex_lock.Lock()
	defer user_mutex_lock.Unlock()
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)
	if user.Uid == 0 {
		uid, err := GetNextUserId()
		if err != nil {
			return errors.As(err, *user)
		}
		user.Uid = uid
	}

	if err := c.Insert(user); err != nil {
		return errors.As(err, *user)
	}

	return nil
}

func GetUserCount() (uint64, error) {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)
	count, err := c.Find(nil).Count()
	if err != nil {
		return 0, errors.As(err)
	}
	return uint64(count), nil
}

func GetNextUserId() (uint64, error) {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)

	userInfo := &UserInfo{}
	if err := c.Find(nil).Sort("-uid").One(&userInfo); err != nil {
		if err == mgo.ErrNotFound {
			return uint64(MIN_USER_ID), nil
		}
		return 0, errors.As(err)
	}
	return userInfo.Uid + 1, nil
}

func DeleteUserInfoByUid(uid uint64) error {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)
	if err := c.Remove(bson.M{"uid": uid}); err != nil {
		return errors.As(err, uid)
	}
	return nil
}

func SetUserInfo(sel, set map[string]interface{}) error {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)
	if err := c.Update(bson.M(sel), bson.M{"$set": bson.M(set)}); err != nil {
		return errors.As(err, sel, set)
	}
	return nil
}
func GetUserInfo(sel map[string]interface{}) (*UserInfo, error) {
	c := common.MongoCollection(USER_DB, USER_INFO_TABLE)

	user := &UserInfo{}
	if err := c.Find(bson.M(sel)).One(user); err != nil {
		return nil, errors.As(err, sel)
	}
	return user, nil
}

func GetUserInfoByUid(uid uint64) (*UserInfo, error) {
	return GetUserInfo(map[string]interface{}{
		"uid": uid,
	})
}

func GetUserInfoByPhoneNum(phonenum string) (*UserInfo, error) {
	return GetUserInfo(map[string]interface{}{
		"phonenum": phonenum,
	})
}

func ModifyUserPwdByUid(uid uint64, pwd string) error {
	sel := map[string]interface{}{"uid": uid}
	set := map[string]interface{}{"password": pwd}
	if err := SetUserInfo(sel, set); err != nil {
		return errors.As(err, uid, pwd)
	}
	return nil
}

func ModifyUserPwdByPhoneNum(phonenum, pwd string) error {
	sel := map[string]interface{}{"phonenum": phonenum}
	set := map[string]interface{}{"password": pwd}
	if err := SetUserInfo(sel, set); err != nil {
		return errors.As(err, phonenum, pwd)
	}
	return nil
}
