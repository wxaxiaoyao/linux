package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/qiniu/api/io"

	_ "sirendaou.com/duserver/common"
)

func TestPutFile(t *testing.T) {
	token, key := GetToken()

	ret := io.PutRet{}
	extra := &io.PutExtra{}

	r := bytes.NewBufferString("hello qiniu")
	if err := io.Put(nil, &ret, token, key, r, extra); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)

	url := GetUrl(key)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func TestDeleteFile(t *testing.T) {
	token, key := GetToken()

	ret := io.PutRet{}
	extra := &io.PutExtra{}

	r := bytes.NewBufferString("hello qiniu")
	if err := io.Put(nil, &ret, token, key, r, extra); err != nil {
		fmt.Println(err)
		return
	}
	SetFileExpiredTime(key, 2)
	SetFileExpiredTime("no_exist_file", 3)

	time.Sleep(time.Second * 5)
}
