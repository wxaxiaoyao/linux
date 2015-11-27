package user

import (
	"encoding/json"
	"fmt"

	"sirendaou.com/duserver/common/errors"
	redis_ "sirendaou.com/duserver/common/redis"
)

type UserState struct {
	Uid      uint64
	Sid      uint32
	ConnIP   int64  //连接服务器IP
	ConnPort uint32 //连接服务器PORT
	Online   bool
	SetupId  string //设备ID
}

func SetUserState(us *UserState) error {
	key := fmt.Sprintf("user_state_%v", us.Uid)
	body, err := json.Marshal(us)
	if err != nil {
		return errors.As(err, *us)
	}
	if err := redis_.Set(key, string(body)); err != nil {
		return errors.As(err, string(body))
	}
	return nil
}

func GetUserState(uid uint64) (*UserState, error) {
	us := &UserState{Uid: uid}
	key := fmt.Sprintf("user_state_%v", us.Uid)
	val, err := redis_.Get(key)
	if err != nil {
		if errors.ERR_NO_DATA.Equal(err) {
			return us, nil
		}
		return us, errors.As(err, key)
	}
	if err := json.Unmarshal([]byte(val), us); err != nil {
		return us, errors.As(err, val)
	}
	return us, nil
}

func SetUserStateOnline(uid uint64, online bool) error {
	us, err := GetUserState(uid)
	if err != nil {
		return errors.As(err, uid, online)
	}
	us.Online = online
	if err := SetUserState(us); err != nil {
		return errors.As(err, uid, online)
	}
	return nil
}

func GetUserStateOnline(uid uint64) bool {
	us, err := GetUserState(uid)
	if err != nil {
		return false
	}
	return us.Online
}
