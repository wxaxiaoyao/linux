package trace

import (
	"fmt"
	"testing"

	"sirendaou.com/duserver/common/redis"
)

var (
	g_trace = New()
)

func test2(args ...interface{}) {
	g_trace.Trace("print trace info")
	fmt.Println(g_trace.String())
}

func test1(args ...interface{}) {
	g_trace.Trace("thi is a trace test", fmt.Sprint(args))
	test2()
}

func TestTrace(t *testing.T) {
	g_trace.Trace("main func")
	test1("hello world")
}

func TestTraceMgr(t *testing.T) {
	redis.Init("127.0.0.1:6379", 1)
	tt := New()
	tt.Trace("test1")
	id := tt.Identify()

	Find(id).Trace("test2")

	fmt.Println(tt.String())

	tt.Save()

	ttt := Load(tt.Identify())

	Trace(ttt.ID, "test3")
	fmt.Println(ttt)
}
