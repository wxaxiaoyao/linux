package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/redis"
)

type TextTeamMsg struct {
	CmdType    uint64 `json:"cmdtype"`
	FromUid    uint64 `json:"fromuid"`
	ToTeamId   uint64 `json:"toteamid"`
	SendTime   int    `json:"sendtime"`
	MsgContent string `json:"msgcontent"`
	MsgId      uint64 `json:"msgid"`
	MsgType    int    `json:"msgtype"`
	ApnsText   string `json:"apnstext,omitempty"`
	FBv        int    `json:"frombv"`
}

func RedisSetTeamMsg(msgId uint64, msg []byte) error {
	userKey := fmt.Sprintf("%s%d", KEY_TEAMMSGBUF, msgId)
	redis.RedisSetEx(userKey, time.Second*3*86400, string(msg[:]))
	return nil
}

func RedisGetTeamMsg(msgId uint64) ([]byte, error) {
	userKey := fmt.Sprintf("%s%d", KEY_TEAMMSGBUF, msgId)
	val, err := redis.RedisGet(userKey)
	if err != nil {
		return nil, errors.As(err, msgId)
	}

	return []byte(val), nil
}

func RedisDelMsgBuf(msgId uint64, msg []byte) error {
	userKey := fmt.Sprintf("%s%d", KEY_TEAMMSGBUF, msgId)
	redis.RedisDel(userKey)
	return nil
}

func RedisAddUserTeamMsg(uids []uint64, msgId uint64, score float64) error {
	for _, uid := range uids {
		if uid <= 100000 {
			continue
		}

		userKey := fmt.Sprintf("%s%d", SET_TEAMMSGID, uid)
		cnt, err := redis.ZCard(userKey).Result()
		if err != nil {
			cnt = 0
		}

		if int(cnt) >= common.MAX_TEAM_MSG_PER {
			redis.ZRemRangeByRank(userKey, 0, 0).Result()
		}

		val := fmt.Sprintf("%d", msgId)
		_, err = redis.ZAdd(userKey, redis.Z{score, val}).Result()
		if err != nil {
			return errors.As(err, uids, msgId, score)
		}
	}

	return nil
}

func RedisDeleteUserTeamMsg(uid, msgId uint64) error {
	userKey := fmt.Sprintf("%s%d", SET_TEAMMSGID, uid)
	member := fmt.Sprintf("%d", msgId)
	redis.RedisZRem(userKey, member)
	return nil
}

func RedisGetUserTeamMsgList(uid uint64) ([]uint64, error) {
	userKey := fmt.Sprintf("%s%d", SET_TEAMMSGID, uid)
	vals, err := redis.RedisZRange(userKey)

	if err != nil {
		return nil, errors.As(err, uid)
	}

	uidList := make([]uint64, len(vals))
	for i, s := range vals {
		uidList[i], err = strconv.ParseUint(s, 10, 64)
	}

	return uidList, nil
}

func GetUserTeamMsgList(uid uint64) ([]*TextTeamMsg, error) {
	msgList := []*TextTeamMsg{}
	msgIdList, err := RedisGetUserTeamMsgList(uid)
	if err != nil {
		return msgList, errors.As(err, uid)
	}
	for _, msgId := range msgIdList {
		msgBuf, err := RedisGetTeamMsg(msgId)
		if err != nil {
			return msgList, errors.As(err, msgId)
		}
		_, jsonBody, _, err := common.UnpackageData(msgBuf)
		if err != nil {
			return msgList, errors.As(err, uid)
		}
		msg := new(TextTeamMsg)
		if err := json.Unmarshal(jsonBody, msg); err != nil {
			return msgList, errors.As(err, string(jsonBody))
		}

		msg.CmdType = common.DU_PUSH_CMD_IM_TEAM_MSG
		msgList = append(msgList, msg)
	}

	return msgList, nil
}
