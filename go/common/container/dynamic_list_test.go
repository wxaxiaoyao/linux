package container

import (
	"fmt"
	"testing"
)

func Interface() interface{} {
	return Type()
}
func Type() *int {
	return nil
}
func TestInterface(t *testing.T) {
	if i := Interface(); i != nil {
		fmt.Println("not nil")
	} else {
		fmt.Println("is nil")
	}
}
