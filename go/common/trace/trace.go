package trace

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"sirendaou.com/duserver/common/redis"
	"sirendaou.com/duserver/common/safemap"
	"sirendaou.com/duserver/common/uuid"
)

var (
	g_traceMgr = safemap.New()
)

type TraceT struct {
	ID     string
	Caller []string
	Info   []string
	Mutex  *sync.Mutex
}

func caller(dept int) string {
	_, file, line, ok := runtime.Caller(dept)
	if !ok {
		return "Unknown"
	}

	idx := strings.LastIndex(file, "/")

	return fmt.Sprint(file[idx+1:], ":", line, " [", time.Now().Format("2006-01-02 15:04:05"), "]")
}

// 新建错误
func New() *TraceT {
	t := &TraceT{
		ID:     uuid.GetUid(),
		Caller: []string{},
		Info:   []string{},
		Mutex:  &sync.Mutex{},
	}
	g_traceMgr.Set(t.ID, t)

	return t
}

func Find(id string) *TraceT {
	t := g_traceMgr.Get(id)
	if t == nil {
		return Load(id)
	}
	return t.(*TraceT)
}

func Delete(id string) {
	g_traceMgr.Delete(id)
	redis.Del(id)
}

func Trace(id string, info ...interface{}) {
	t := Find(id)
	if t == nil {
		return
	}
	t.TraceDept(3, info...)
	return
}

func Load(id string) *TraceT {
	body, err := redis.Get(id)
	if err != nil {
		return nil
	}
	t := &TraceT{}
	if err := json.Unmarshal([]byte(body), t); err != nil {
		return nil
	}
	g_traceMgr.Set(t.ID, t)
	return t
}

func Save(id string) {
	t := Find(id)
	if t == nil {
		return
	}
	t.Save()
	return
}

func String(id string) string {
	t := Find(id)
	if t == nil {
		return "Not Found"
	}
	return t.String()
}

// package
func (t *TraceT) Save() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	body, err := json.Marshal(t)
	if err != nil {
		println(err.Error())
		return
	}
	if err := redis.SetEx(t.ID, time.Minute*5, string(body)); err != nil {
		println(err.Error())
		return
	}
	return
}

// trace ID
func (t *TraceT) Identify() string {
	return t.ID
}

// 追加错误
func (t *TraceT) Trace(info ...interface{}) *TraceT {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Caller = append(t.Caller, caller(2))
	t.Info = append(t.Info, fmt.Sprint(info))
	return t
}

func (t *TraceT) TraceDept(dept int, info ...interface{}) *TraceT {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Caller = append(t.Caller, caller(dept))
	t.Info = append(t.Info, fmt.Sprint(info))
	return t
}

// 复制
func (t *TraceT) Clone() *TraceT {
	return &TraceT{
		Caller: append([]string{}, t.Caller...),
		Info:   append([]string{}, t.Info...),
	}
}

// error 接口实现
func (t *TraceT) String() string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	count := len(t.Caller)
	traceStr := fmt.Sprint("\ntrace_info:")
	for i := 0; i < count; i++ {
		traceStr += fmt.Sprintf("\ntrace_%v==> caller:%v info:%v", i, t.Caller[i], t.Info[i])
	}
	return traceStr
}

// 清空
func (t *TraceT) Clear() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Caller = []string{}
	t.Info = []string{}
}
