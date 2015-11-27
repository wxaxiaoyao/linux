package user

import (
	"fmt"
	"strconv"
	"time"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/redis"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Friend struct {
	Uid      uint64
	FUid     uint64
	Type     int
	CreateDT string
	UpdateDT string
}

func FriendInit() error {
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	index := mgo.Index{
		Key:      []string{"uid", "fuid"},
		Unique:   true,
		DropDups: true,
	}

	return errors.As(c.EnsureIndex(index))
}

func AddFriend(friend *Friend) error {
	//sel := map[string]interface{}{"uid":friend.Uid, "fuid":friend.FUid}
	//set := map[string]interface{}{"type":friend.Type}
	if err := DeleteFriend(friend.Uid, friend.FUid); err != nil {
		return errors.As(err, *friend)
	}
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	friend.CreateDT = time.Now().Format(common.DATETIME_FMT)
	friend.UpdateDT = time.Now().Format(common.DATETIME_FMT)
	return errors.As(c.Insert(friend), *friend)
}

func DeleteFriend(uid, fid uint64) error {
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	err := c.Remove(bson.M{"fuid": fid, "uid": uid})
	if err != mgo.ErrNotFound {
		return errors.As(err, fid)
	}
	return nil
}

func GetFriend(uid, fid uint64) *Friend {
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	friend := &Friend{}

	iter := c.Find(bson.M{"fuid": fid, "uid": uid}).Iter()
	defer iter.Close()

	if iter.Next(friend) {
		return friend
	}
	return nil
}

func GetFriendList(uid uint64, typ int) ([]uint64, error) {
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	friend := &Friend{}
	uids := []uint64{}
	iter := c.Find(bson.M{"uid": uid, "type": typ}).Iter()
	defer iter.Close()

	for iter.Next(friend) {
		uids = append(uids, friend.FUid)
	}
	return uids, nil
}

func SetFriend(sel, set map[string]interface{}) error {
	c := common.MongoCollection(USER_DB, USER_FRIEND_TABLE)

	if err := c.Update(bson.M(sel), bson.M{"$set": bson.M(set)}); err != nil {
		return errors.As(err, sel, set)
	}
	return nil
}

func RedisAddFriend(uid, fid uint64, typ int) error {
	userKey := ""
	if err := RedisDeleteFriend(uid, fid, typ); err != nil {
		return errors.As(err, uid, fid, typ)
	}

	if typ == 1 {
		userKey = fmt.Sprintf("%s%d", SET_WHITELIST, uid)
	} else {
		userKey = fmt.Sprintf("%s%d", SET_BLACKLIST, uid)
		// 加入黑名单 删除白名单
		if err := RedisDeleteFriend(uid, fid, 1); err != nil {
			return errors.As(err, uid, fid, typ)
		}
	}

	val := fmt.Sprintf("%d", fid)
	redis.RedisSAdd(userKey, val)

	// 双向好友
	if typ == 1 {
		key_2 := fmt.Sprintf("%s%d", SET_WHITELIST, fid)
		uid_2 := fmt.Sprintf("%d", uid)
		redis.RedisSAdd(key_2, uid_2)
	}

	return nil
}

func RedisDeleteFriend(uid, fid uint64, typ int) error {
	userKey := ""
	if typ == 1 {
		userKey = fmt.Sprintf("%s%d", SET_WHITELIST, uid)
	} else {
		userKey = fmt.Sprintf("%s%d", SET_BLACKLIST, uid)
	}
	uid_1 := fmt.Sprintf("%d", fid)

	redis.SRem(userKey, uid_1)

	// 双向好友
	if typ == 1 {
		userKey = fmt.Sprintf("%s%d", SET_WHITELIST, fid)
		uid_2 := fmt.Sprintf("%d", uid)
		redis.SRem(userKey, uid_2)
	}

	return nil
}

func RedisQueryFriendList(uid uint64, typ int) ([]uint64, error) {
	userKey := ""
	if typ == 1 {
		userKey = fmt.Sprintf("%s%d", SET_WHITELIST, uid)
	} else {
		userKey = fmt.Sprintf("%s%d", SET_BLACKLIST, uid)
	}

	strUidList, err := redis.SMembers(userKey).Result()
	if err != nil {
		return []uint64{}, errors.As(err, uid, typ)
	}

	if strUidList == nil || len(strUidList) < 1 {
		return []uint64{}, nil
	}

	uidList := make([]uint64, len(strUidList))

	for i, val := range strUidList {
		uidList[i], _ = strconv.ParseUint(val, 10, 64)
	}

	return uidList, nil
}

func RedisQueryFriend(uid, fid uint64, typ int) bool {
	userKey := ""
	if typ == 1 {
		userKey = fmt.Sprintf("%s%d", SET_WHITELIST, uid)
	} else {
		userKey = fmt.Sprintf("%s%d", SET_BLACKLIST, uid)
	}

	val := fmt.Sprintf("%d", fid)

	result := redis.SIsMember(userKey, val)
	if result == nil {
		if typ == 1 {
			return true
		} else {
			return false
		}
	}
	return result.Val()
}
