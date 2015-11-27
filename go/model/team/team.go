package team

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/redis"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TEAM_TYPE_INVALID = -1
	TEAM_TYPE_USER    = 0
	TEAM_TYPE_SYS     = 1
)

type TeamInfo struct {
	TeamId     uint64 `json:"teamid,omitempty"`
	Uid        uint64 `json:"uid,omitempty"`
	TeamType   int    `json:"teamtype,omitempty"`
	TeamName   string `json:"teamname,omitempty"`
	CreateDate string `json:"createdate,omitempty"`
	//CreateDate time.Time `json:"createdate,omitempty"`
	CoreInfo string `json:"coreinfo,omitempty"`
	ExInfo   string `json:"exinfo,omitempty"`
	MaxCount int    `json:"maxcount,omitempty"`
	IV       int64  `json:"infov,omitempty"`
	MV       int64  `json:"memberv,omitempty"`
}

func TeamInfoInit() error {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)

	index := mgo.Index{
		Key:      []string{"teamid"},
		Unique:   true,
		DropDups: true,
	}

	return errors.As(c.EnsureIndex(index))
}

func CreateTeam(team *TeamInfo) error {
	team_mutex_lock.Lock()
	defer team_mutex_lock.Unlock()

	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)
	if team.TeamId == 0 {
		teamId, err := GetNextTeamId(team.Uid)
		if err != nil {
			return errors.As(err, *team)
		}
		team.TeamId = teamId
	}

	team.CreateDate = time.Now().Format(DATETIME_FMT)
	return errors.As(c.Insert(team), *team)
}

func DeleteTeam(teamid uint64) error {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)

	err := c.Remove(bson.M{"teamid": teamid})
	if err != mgo.ErrNotFound {
		return errors.As(err, teamid)
	}
	return nil
}

func GetTeam(teamid uint64) *TeamInfo {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)

	team := &TeamInfo{}

	iter := c.Find(bson.M{"teamid": teamid}).Iter()
	defer iter.Close()

	if iter.Next(team) {
		return team
	}
	return nil
}

func GetTeamList(sel map[string]interface{}, start, count int) ([]TeamInfo, error) {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)

	team_list := []TeamInfo{}
	err := c.Find(bson.M(sel)).Skip(start).Limit(count).All(&team_list)

	return team_list, errors.As(err, sel)
}

func GetSysTeamList() ([]TeamInfo, error) {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)
	teamList := []TeamInfo{}
	team := &TeamInfo{}

	iter := c.Find(bson.M{"teamtype": 1}).Iter()
	defer iter.Close()

	for iter.Next(team) {
		teamList = append(teamList, *team)
	}
	return teamList, nil

}

func SetTeam(sel, set map[string]interface{}) error {
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)

	if err := c.Update(bson.M(sel), bson.M{"$set": bson.M(set)}); err != nil {
		return errors.As(err, sel, set)
	}
	return nil
}

func GetNextTeamId(uid uint64) (uint64, error) {
	// lock uid
	c := common.MongoCollection(TEAM_DB, TEAM_INFO_TABLE)
	ti := &TeamInfo{}
	// 先找一个无效的组id，相当于被删除的组
	if err := c.Find(bson.M{"uid": uid, "teamtype": TEAM_TYPE_INVALID}).One(ti); err != nil && err != mgo.ErrNotFound {
		return 0, errors.As(err, uid)
	}
	// 若没有无效组，查找已有最大组id
	if err := c.Find(bson.M{"uid": uid}).Sort("-teamid").One(ti); err != nil {
		if err == mgo.ErrNotFound {
			// 没有组，从1计数
			return uint64(int(uid)*MAX_NUM_TEAM + 1), nil
		}
		return 0, errors.As(err, uid)
	}
	// 判断是否超过用户组上限
	if int(ti.TeamId)%MAX_NUM_TEAM >= MAX_NUM_TEAM_USER {
		return 0, errors.New("user create team too many", uid)
	}

	return ti.TeamId + 1, nil
}

func IsTeamCreator(uid, tid uint64) bool {
	return (tid / uint64(MAX_NUM_TEAM)) == uid
}

func RedisCreateTeam(team *TeamInfo) error {
	b, err := json.Marshal(*team)
	if err != nil {
		return errors.As(err, *team)
	}

	userKey := fmt.Sprintf("%s%d", KEY_TEAM_INFO, team.TeamId)
	redis.RedisSet(userKey, string(b))

	userKey = fmt.Sprintf("%s%d", SET_USERS_TEAM, team.Uid)
	val := fmt.Sprintf("%d", team.TeamId)

	redis.RedisSAdd(userKey, val)

	//info version
	val = fmt.Sprintf("%d", time.Now().Unix())
	userKey = fmt.Sprintf("%s%d", KEY_TEAM_INFO_VER, team.TeamId)
	redis.RedisSet(userKey, val)

	return nil
}

func RedisDeleteTeam(tid uint64) error {
	userKey := fmt.Sprintf("%s%d", KEY_TEAM_INFO, tid)
	redis.RedisDel(userKey)

	userKey = fmt.Sprintf("%s%d", SET_TEAM_MEMBER, tid)

	uidList, err := redis.SMembers(userKey).Result()

	if err != nil {
		return errors.As(err, tid)
	} else {
		for _, val := range uidList {
			userKey2 := fmt.Sprintf("%s%s", SET_USERS_TEAM, val)
			val := fmt.Sprintf("%d", tid)
			redis.SRem(userKey2, val)
		}
	}

	redis.RedisDel(userKey)

	return nil
}

func RedisSetTeam(team *TeamInfo) error {
	userKey := fmt.Sprintf("%s%d", KEY_TEAM_INFO, team.TeamId)
	val, err := redis.RedisGet(userKey)
	if err != nil {
		return errors.As(err, *team)
	}

	var t TeamInfo
	err = json.Unmarshal([]byte(val), &t)
	if err != nil {
		return errors.As(err, *team)
	}

	if len(team.TeamName) > 1 {
		t.TeamName = team.TeamName
	}
	if len(team.CoreInfo) > 1 {
		t.CoreInfo = team.CoreInfo
	}

	if len(team.ExInfo) > 1 {
		t.ExInfo = team.ExInfo
	}

	b, err := json.Marshal(t)
	if err != nil {
		return errors.As(err, t)
	}
	userKey = fmt.Sprintf("%s%d", KEY_TEAM_INFO, team.TeamId)
	redis.RedisSet(userKey, string(b))

	//info version
	val = fmt.Sprintf("%d", time.Now().Unix())
	userKey = fmt.Sprintf("%s%d", KEY_TEAM_INFO_VER, team.TeamId)
	redis.RedisSet(userKey, val)

	return nil
}

func RedisQueryTeam(tid uint64) (*TeamInfo, error) {
	userKey := fmt.Sprintf("%s%d", KEY_TEAM_INFO, tid)
	val, err := redis.RedisGet(userKey)
	if err != nil {
		return nil, errors.As(err, tid)
	}

	team := &TeamInfo{}
	if err := json.Unmarshal([]byte(val), team); err != nil {
		return nil, errors.As(err, string(val))
	}

	return team, nil
}

func RedisQueryTeamList(uid uint64) ([]int64, error) {
	userKey := fmt.Sprintf("%s%d", SET_USERS_TEAM, uid)
	strTeamList, err := redis.SMembers(userKey).Result()

	if err != nil {
		return nil, errors.As(err, uid)
	}

	teamList := make([]int64, len(strTeamList))

	for i, val := range strTeamList {
		teamList[i], _ = strconv.ParseInt(val, 10, 64)
	}

	return teamList, nil
}

// 系统预设的群组
func RedisQuerySysTeamList() ([]int64, error) {
	strTeamList, err := redis.SMembers(SET_SYS_TEAM).Result()
	if err != nil {
		return nil, errors.As(err)
	}

	teamList := make([]int64, len(strTeamList))

	for i, val := range strTeamList {
		teamList[i], _ = strconv.ParseInt(val, 10, 64)
	}

	return teamList, nil
}
