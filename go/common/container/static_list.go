package container

import (
	"sirendaou.com/duserver/common/pool"
	"sync"
)

type Node struct {
	Value interface{}
	Next  *Node
	Prev  *Node
}

type StaticList struct {
	*sync.Mutex
	Head *Node
	Pool *pool.Pools
	Size int
}

func NewStaticList() *StaticList {
	list := &StaticList{
		Mutex: &sync.Mutex{},
		Head:  &Node{Value: nil, Next: nil, Prev: nil},
		Pool:  pool.NewPools(Node{}),
		Size:  0,
	}
	list.Head.Next = list.Head
	list.Head.Prev = list.Head

	return list
}

func (this *StaticList) PushFront(v interface{}) {
	node := this.Pool.Get().(*Node)
	node.Value = v

	this.Lock()
	defer this.Unlock()

	node.Next = this.Head.Next
	this.Head.Prev = node
	this.Head.Next = node
	node.Prev = this.Head

	this.Size++
}

func (this *StaticList) PushBack(v interface{}) {
	node := this.Pool.Get().(*Node)
	node.Value = v

	this.Lock()
	defer this.Unlock()

	this.Head.Prev.Next = node
	node.Prev = this.Head.Prev
	this.Head.Prev = node
	node.Next = this.Head

	this.Size++
}

func (this *StaticList) Empty() bool {
	this.Lock()
	defer this.Unlock()
	return this.Size == 0
}

func (this *StaticList) Len() int {
	this.Lock()
	defer this.Unlock()
	return this.Size
}

func (this *StaticList) PopFront() interface{} {
	if this.Empty() {
		return nil
	}

	this.Lock()
	node := this.Head.Next
	this.Head.Next = node.Next
	node.Next.Prev = this.Head
	this.Size--
	this.Unlock()

	value := node.Value
	this.Pool.Put(node)

	return value
}

func (this *StaticList) PopBack() interface{} {
	if this.Empty() {
		return nil
	}

	this.Lock()
	node := this.Head.Prev
	this.Head.Prev = node.Prev
	node.Prev.Next = this.Head
	this.Size--
	this.Unlock()

	value := node.Value
	this.Pool.Put(node)

	return value
}
