package safemap

import (
	"reflect"
	"sync"
	"time"
)

var (
	DEFAULT_TIMEOUT = 5 * time.Minute

	g_safemap *SafeMap = New()
)

type Value struct {
	value interface{}
	timer *time.Timer
}

type SafeMap struct {
	data    map[interface{}]*Value
	rwMutex *sync.RWMutex
}

func New() *SafeMap {
	return &SafeMap{
		data:    make(map[interface{}]*Value),
		rwMutex: &sync.RWMutex{},
	}
}

func Set(key, value interface{}) {
	g_safemap.Set(key, value)
}

func SetByTimeout(key, value interface{}, timeout time.Duration) {
	g_safemap.SetByTimeout(key, value, timeout)
}

func Get(key interface{}) interface{} {
	return g_safemap.Get(key)
}

func Delete(key interface{}) {
	g_safemap.Delete(key)
}

func (this *SafeMap) Set(key, value interface{}) {
	this.SetByTimeout(key, value, DEFAULT_TIMEOUT)
}

func (this *SafeMap) SetByTimeout(key, value interface{}, timeout time.Duration) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	val, ok := this.data[key]
	if ok {
		val.timer.Stop()
	} else {
		val = new(Value)
	}
	val.value = value
	val.timer = time.AfterFunc(timeout, func() {
		this.rwMutex.Lock()
		defer this.rwMutex.Unlock()
		ind := reflect.ValueOf(value)
		if f := ind.MethodByName("SafeMapTimeoutCall"); f.IsValid() {
			f.Call([]reflect.Value{reflect.ValueOf(key)})
		}
		delete(this.data, key)
	})
	this.data[key] = val
}

func (this *SafeMap) Get(key interface{}) interface{} {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	val, ok := this.data[key]
	if ok {
		return val.value
	}
	return nil
}

func (this *SafeMap) Delete(key interface{}) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	delete(this.data, key)
}

func (this *SafeMap) Foreach(key interface{}, comp func(key, mapKey interface{}) bool) []interface{} {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()

	valSlice := []interface{}{}
	for k, v := range this.data {
		if comp(key, k) {
			valSlice = append(valSlice, v.value)
		}
	}
	return valSlice
}
