package uuid

import (
	"fmt"
	"reflect"
	"testing"
)

type T struct {
	s []byte
}

func TestUuid(tt *testing.T) {
	id := GetUid()
	t := &T{}
	t.s = []byte(id)
	var a [12]byte

	var i interface{} = a
	switch i.(type) {
	case []byte:
		fmt.Println("good")
	}
	v := reflect.ValueOf(t.s)
	fmt.Println(v, len(t.s))
}
