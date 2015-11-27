package team

import (
	"sync"

	"sirendaou.com/duserver/common/errors"
)

var (
	DATETIME_FMT        = "2006-01-02 15:04:05"
	MAX_MEMBER_NUM_TEAM = 500
	MAX_NUM_TEAM        = 1000
	MAX_NUM_TEAM_USER   = 300

	TEAM_DB           = "dudb"
	TEAM_INFO_TABLE   = "team_info"
	TEAM_MEMBER_TABLE = "team_member"

	SET_USERS_TEAM      = "userteam_"
	SET_SYS_TEAM        = "systeam_"
	SET_TEAM_MEMBER     = "teammember_"
	KEY_TEAM_INFO       = "teaminfo_"
	KEY_TEAM_INFO_VER   = "tinfov_"
	KEY_TEAM_MEMBER_VER = "tmemberv_"

	team_mutex_lock = &sync.Mutex{}
)

func init() {
	if err := TeamMemberInit(); err != nil {
		panic(errors.As(err).Error())
	}

	if err := TeamInfoInit(); err != nil {
		panic(errors.As(err).Error())
	}
}
