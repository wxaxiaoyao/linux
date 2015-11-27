package user

import (
	"fmt"
	"testing"
	"time"

	"sirendaou.com/duserver/common"
)

func TestPut(t *testing.T) {
	common.MysqlInit("127.0.0.1", "testDB", "root", "root")
	user := &UserInfo{
		Uid:      100000,
		PhoneNum: "18702759796",
		Password: "test",
	}
	DeleteUserInfoByUid(user.Uid)
	if err := SaveUserInfo(user); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(user.Uid)

	time.Sleep(time.Second)
	if err := ModifyUserPwdByUid(user.Uid, "123"); err != nil {
		fmt.Println(err)
		return
	}

	tmpUser, err := GetUserInfoByUid(user.Uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*tmpUser)

	if err := ModifyUserPwdByPhoneNum(user.PhoneNum, "12a3"); err != nil {
		fmt.Println(err)
		return
	}
	tmpUser, err = GetUserInfoByPhoneNum(user.PhoneNum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*tmpUser)

}
