package message

import (
	"sirendaou.com/duserver/common/errors"
)

var (
	MSG_DB                = "dudb"
	MSG_USER_MSG_TABLE    = "user_msg"
	MSG_TEAM_INFO_TABLE   = "team_info"
	MSG_TEAM_MEMBER_TABLE = "team_table"

	KEY_TEAMMSGBUF = "TEAMMSGBUF_"
	SET_TEAMMSGID  = "TEAMMSGID_"
)

const (
	MSG_TYPE_USER = iota
	MSG_TYPE_TEAM
	MSG_TYPE_SYSTEM
)

func init() {
	if err := UserMsgInit(); err != nil {
		panic(errors.As(err).Error())
	}
}
