package user

import (
	"fmt"
	"sirendaou.com/duserver/common/redis"
)

func SetUserSetupId(uid uint64, setupId string) error {
	return redis.Set(fmt.Sprint(USER_TOKEN, uid), setupId)
}

func GetUserSetupId(uid uint64) (string, error) {
	return redis.Get(fmt.Sprint(USER_TOKEN, uid))
}

func IsSameSetupId(uid uint64, setupid string) bool {
	id, _ := GetUserSetupId(uid)
	if id == setupid {
		return true
	}

	return false
}
