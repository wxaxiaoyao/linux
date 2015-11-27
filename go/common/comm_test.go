package common

import (
	"fmt"
	"testing"
)

func TestTailZero(t *testing.T) {
	tail := &InnerPkgTail{
		TailID: TailID(""),
	}
	fmt.Println(tail.TailId())
}
