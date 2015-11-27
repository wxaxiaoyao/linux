package errors

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	/*
			err := New("100", "test", "test1")
			fmt.Println(err)

			err.As("test2")
			fmt.Println(err.Error())

		e1 := errors.New("hello world")
		e2 := As(e1, "test1")
		fmt.Println(e2)

		e3 := As(e2, "test2")
		fmt.Println(e3)
	*/
	var e1 error
	e1 = As(nil, "test")
	if e1 != nil {
		fmt.Println(e1)
	}

}
