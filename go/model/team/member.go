package team

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

type TeamMember struct {
	TeamId uint64
	Uid    uint64
	Date   string
	//nickname string
}

func TeamMemberInit() error {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	index := mgo.Index{
		Key:      []string{"teamid", "uid"},
		Unique:   true,
		DropDups: true,
	}

	return errors.As(c.EnsureIndex(index))
}
func AddTeamMember(tm *TeamMember) error {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	tm.Date = time.Now().Format(DATETIME_FMT)
	return errors.As(c.Insert(tm), *tm)
}

func DeleteTeamMember(teamid, uid uint64) error {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	err := c.Remove(bson.M{"teamid": teamid, "uid": uid})
	if err != mgo.ErrNotFound {
		return errors.As(err, uid, teamid)
	}
	return nil
}

func GetTeamMember(teamid, uid uint64) *TeamMember {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	tm := &TeamMember{}
	iter := c.Find(bson.M{"teamid": teamid, "uid": uid}).Iter()
	defer iter.Close()

	if iter.Next(tm) {
		return tm
	}

	return nil
}

func GetTeamMemberList(teamid uint64) ([]uint64, error) {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	tm := &TeamMember{}
	uids := []uint64{}

	iter := c.Find(bson.M{"teamid": teamid}).Iter()
	defer iter.Close()

	for iter.Next(tm) {
		uids = append(uids, tm.Uid)
	}

	return uids, nil
}

func SetTeamMember(sel, set map[string]interface{}) error {
	c := common.MongoCollection(TEAM_DB, TEAM_MEMBER_TABLE)

	if err := c.Update(bson.M(sel), bson.M{"$set": bson.M(set)}); err != nil {
		return errors.As(err, sel, set)
	}
	return nil
}

func RedisQueryTeamMemberList(tid uint64) ([]uint64, error) {
	userKey := fmt.Sprintf("%s%d", SET_TEAM_MEMBER, tid)
	strUidList, err := redis.SMembers(userKey).Result()
	if err != nil {
		return nil, errors.As(err, tid)
	}

	uidList := make([]uint64, len(strUidList))

	for i, val := range strUidList {
		uidList[i], _ = strconv.ParseUint(val, 10, 64)
	}

	return uidList, nil
}

func RedisIsTeamMember(teamId, uid uint64) (bool, error) {
	userKey := fmt.Sprintf("%s%d", SET_TEAM_MEMBER, teamId)
	val := fmt.Sprintf("%d", uid)

	is, err := redis.SIsMember(userKey, val).Result()

	if err != nil {
		return false, errors.As(err, teamId, uid)
	}

	return is, nil
}

func RedisAddTeamMember(tid, uid uint64) error {
	count := RedisTeamMemberCount(tid)
	if int(count) >= MAX_MEMBER_NUM_TEAM {
		return errors.New("team member too many")
	}
	userKey := fmt.Sprintf("%s%d", SET_TEAM_MEMBER, tid)
	val := fmt.Sprintf("%d", uid)

	redis.RedisSAdd(userKey, val)

	//member version
	val = fmt.Sprintf("%d", time.Now().Unix())
	userKey = fmt.Sprintf("%s%d", KEY_TEAM_MEMBER_VER, tid)
	redis.RedisSet(userKey, val)

	userKey = fmt.Sprintf("%s%d", SET_USERS_TEAM, uid)
	val = fmt.Sprintf("%d", tid)
	redis.RedisSAdd(userKey, val)

	return nil
}

func RedisTeamMemberCount(tid uint64) int64 {
	userKey := fmt.Sprintf("%s%d", SET_TEAM_MEMBER, tid)

	num := redis.SCard(userKey).Val()

	return num
}

func RedisDeleteTeamMember(tid, uid uint64) error {
	userKey := fmt.Sprintf("%s%d", SET_TEAM_MEMBER, tid)
	val := fmt.Sprintf("%d", uid)
	redis.SRem(userKey, val)

	//member version
	val = fmt.Sprintf("%d", time.Now().Unix())
	userKey = fmt.Sprintf("%s%d", KEY_TEAM_MEMBER_VER, tid)
	redis.RedisSet(userKey, val)

	userKey = fmt.Sprintf("%s%d", SET_USERS_TEAM, uid)
	val = fmt.Sprintf("%d", tid)
	redis.SRem(userKey, val)

	return nil
}
