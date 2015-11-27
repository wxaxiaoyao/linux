package message

import (
	"fmt"
	"testing"

	"sirendaou.com/duserver/common"
)

func TestUserMsg(t *testing.T) {
	common.MongoInit("127.0.0.1")
	if err := UserMsgInit(); err != nil {
		fmt.Println(err)
		return
	}
	msg := &UserMsgItem{
		MsgId:   1,
		FromUid: 2,
		ToUid:   3,
		Content: "hello world",
	}
	DelUserMsg(msg.ToUid, msg.MsgId)
	if err := SaveUserMsg(msg); err != nil {
		fmt.Println(err)
		return
	}

	msgs, _ := GetUserMsgList(msg.ToUid)

	for _, m := range msgs {
		fmt.Println(*m)
	}
	fmt.Println("msg count:", len(msgs))
}
