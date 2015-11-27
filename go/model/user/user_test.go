package user

import (
	"fmt"
	"testing"

	"sirendaou.com/duserver/common"
)

func TestUserInfo(t *testing.T) {
	if err := UserInfoInit(); err != nil {
		fmt.Println(err)
		return
	}

	DeleteUserInfoByUid(uint64(MIN_USER_ID))
	DeleteUserInfoByUid(uint64(MIN_USER_ID + 1))
	DeleteUserInfoByUid(uint64(MIN_USER_ID + 2))

	user := &UserInfo{
		Uid:      0,
		Password: "test",
		PhoneNum: "18702759796",
		Platform: "a",
		DeviceId: "123456789",
	}
	if err := SaveUserInfo(user); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*user)
	user.Uid = 0
	if err := SaveUserInfo(user); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*user)
	user.Uid = 0
	if err := SaveUserInfo(user); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*user)

	tmp, err := GetUserInfoByUid(user.Uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*tmp)

	if err := SetUserInfo(
		map[string]interface{}{"uid": user.Uid},
		map[string]interface{}{"platform": "i"},
	); err != nil {
		fmt.Println(err)
		return
	}

	tmp, err = GetUserInfoByUid(user.Uid)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*tmp)
}

func TestBatCreateUser(t *testing.T) {
	user := &UserInfo{Password: "test"}

	for i := 0; i < 1000000; i++ {
		user.Uid = uint64(1000000 + i)
		if err := SaveUserInfo(user); err != nil {
			fmt.Println(err)
		}
	}
}

func TestBatDeleteUser(t *testing.T) {
	for i := 0; i < 1000000; i++ {
		DeleteUserInfoByUid(uint64(1000000 + i))
	}
}
func TestMysqlToMongo(t *testing.T) {
	UserInfoInit()
	sqlStr := "select uid, phonenum, password, platform, did, setupid,baseinfo, exinfo from t_user_info"
	row, err := common.MysqlQuery(sqlStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer row.Close()
	userInfo := &UserInfo{}
	for row.Next() {
		if err := row.Scan(&userInfo.Uid,
			&userInfo.PhoneNum,
			&userInfo.Password,
			&userInfo.Platform,
			&userInfo.DeviceId,
			&userInfo.SetupId,
			&userInfo.BaseInfo,
			&userInfo.ExInfo); err != nil {
			fmt.Println(err)
		}

		if err := SaveUserInfo(userInfo); err != nil {
			fmt.Println(err)
		}
	}
}
