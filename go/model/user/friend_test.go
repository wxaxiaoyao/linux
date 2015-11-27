package user

import (
	"fmt"
	"testing"

	_ "sirendaou.com/duserver/common"
)

func TestFriend(t *testing.T) {
	fmt.Println(RedisQueryFriend(100000, 100001, 2))
}
