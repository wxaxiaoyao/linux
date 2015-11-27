package container

import (
	"container/list"
	"sync"
)

type DynamicList struct {
	List  *list.List
	Mutex *sync.Mutex
}

func NewDynamicList() *DynamicList {
	return &DynamicList{
		List:  list.New(),
		Mutex: &sync.Mutex{},
	}
}

func (this *DynamicList) PopBack() interface{} {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	e := this.List.Back()
	if e == nil {
		return nil
	}
	this.List.Remove(e)
	return e.Value
}

func (this *DynamicList) PopFront() interface{} {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	e := this.List.Front()
	if e == nil {
		return nil
	}
	this.List.Remove(e)
	return e.Value
}

func (this *DynamicList) PushBack(v interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.List.PushBack(v)
	return
}

func (this *DynamicList) PushFront(v interface{}) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.List.PushFront(v)
	return
}

func (this *DynamicList) Len() int {
	this.Mutex.Lock()
	this.Mutex.Unlock()
	return this.List.Len()
}
