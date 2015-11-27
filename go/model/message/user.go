package message

import (
	"fmt"
	"strings"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
	"sirendaou.com/duserver/common/safemap"
	"sirendaou.com/duserver/common/syslog"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserMsgItem struct {
	MsgId    uint64 `json:"msgid"`
	FromUid  uint64 `json:"fromuid"`
	ToUid    uint64 `json:"touid"`
	Type     uint16 `json:"msgtype"`
	Content  string `json:"msgcontent"`
	SendTime uint32 `json:"sendtime"`
	// invalid connect
	CmdType   uint64 `json:"cmdtype"`
	ApnsText  string `json:"apnstext,omitempty"`
	FType     int    `json:"ftype,omitempty"`
	FBv       int    `json:"frombv,omitempty"`
	ExtraData string `json:"extraData"`
}

type UserMsgMgr struct {
	msgMap *safemap.SafeMap
}

var (
	g_userMsgMgr *UserMsgMgr = nil
)

func init() {
	g_userMsgMgr = &UserMsgMgr{
		msgMap: safemap.New(),
	}
}

func msgKey(uid, msgid uint64) string {
	return fmt.Sprintf("%v-%v", uid, msgid)
}

func keyComp(k1, k2 interface{}) bool {
	s1, ok1 := k1.(string)
	s2, ok2 := k2.(string)
	if ok1 && ok2 && strings.HasPrefix(s2, s1) {
		return true
	}
	return false
}

func UserMsgInit() error {
	c := common.MongoCollection(MSG_DB, MSG_USER_MSG_TABLE)

	index := mgo.Index{
		Key: []string{"touid", "msgid"},
	}

	return errors.As(c.EnsureIndex(index))
}

func SaveUserMsg(userMsg *UserMsgItem) error {
	g_userMsgMgr.msgMap.Set(msgKey(userMsg.ToUid, userMsg.MsgId), userMsg)
	return nil
}

func DelUserMsg(uid, msgid uint64) error {
	// 在缓存就不操作数据库
	key := msgKey(uid, msgid)
	if msg := g_userMsgMgr.msgMap.Get(key); msg != nil {
		g_userMsgMgr.msgMap.Delete(key)
		return nil
	}
	// 不在缓存就删除数据库
	g_userMsgMgr.msgMap.Delete(msgKey(uid, msgid))

	c := common.MongoCollection(MSG_DB, MSG_USER_MSG_TABLE)

	err := c.Remove(bson.M{"msgid": msgid, "touid": uid})
	if err != mgo.ErrNotFound {
		return errors.As(err, uid, msgid)
	}
	return nil
}

func GetUserMsgList(uid uint64) ([]*UserMsgItem, error) {
	msgs := []*UserMsgItem{}

	key := fmt.Sprintf("%v-", uid)
	ms := g_userMsgMgr.msgMap.Foreach(key, keyComp)
	for _, v := range ms {
		m := v.(*UserMsgItem)
		msgs = append(msgs, m)
	}

	c := common.MongoCollection(MSG_DB, MSG_USER_MSG_TABLE)
	iter := c.Find(bson.M{"touid": uid}).Iter()
	defer iter.Close()
	msg := new(UserMsgItem)
	for iter.Next(msg) {
		msgs = append(msgs, msg)
		msg = new(UserMsgItem)
	}

	return msgs, nil
}

func (msg *UserMsgItem) SafeMapTimeoutCall(key interface{}) {
	// 超时未删除则保存到db
	c := common.MongoCollection(MSG_DB, MSG_USER_MSG_TABLE)
	if err := c.Insert(msg); err != nil {
		syslog.Info(err, *msg)
	}
}
