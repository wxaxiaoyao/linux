package user

import (
	"fmt"
	"sirendaou.com/duserver/common/redis"
)

func SetUserToken(uid uint64, token string) error {
	return redis.Set(fmt.Sprint(USER_TOKEN, uid), token)
}

func GetUserToken(uid uint64) (string, error) {
	return redis.Get(fmt.Sprint(USER_TOKEN, uid))
}
