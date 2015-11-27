package safemap

import (
	"fmt"
	"log"
	"testing"
	"time"
)

type T struct{}

func (*T) SafeMapTimeoutCall(key interface{}) {
	fmt.Println("hello:", key)
}

func TestSafeMap(t *testing.T) {
	tt := &T{}
	SetByTimeout("test", "hello world", time.Second)
	SetByTimeout(1, "test1", time.Second)
	SetByTimeout(2, tt, time.Second)

	if val := Get("test"); val != nil {
		value := val.(string)
		log.Println(value)
	}

	if val := Get(1); val != nil {
		value := val.(string)
		log.Println(value)
	}

	time.Sleep(time.Second)

	if val := Get(1); val != nil {
		value := val.(string)
		log.Println(value)
	} else {
		log.Println("delete")
	}
}
