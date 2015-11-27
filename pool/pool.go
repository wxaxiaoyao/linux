package pool

import (
	"reflect"
	"sync"
)

var (
	DEFAULT_OBJECT_SIZE = 10000
)

type Pool struct {
	*sync.Mutex
	typ       reflect.Type
	data      reflect.Value
	addrIndex map[interface{}]int
	next      []int
	head      int
	size      int
}

func NewPool(obj interface{}, size int) *Pool {
	ind := reflect.ValueOf(obj)
	typ := reflect.Indirect(ind).Type()
	if typ.Kind() == reflect.Chan || typ.Kind() == reflect.Map || typ.Kind() == reflect.Slice {
		println("not support chan and map slice!!!")
		return nil
	}

	pool := &Pool{
		Mutex:     &sync.Mutex{},
		typ:       typ,
		data:      reflect.MakeSlice(reflect.SliceOf(typ), size, size),
		addrIndex: make(map[interface{}]int),
		next:      make([]int, size, size),
		head:      0,
		size:      size,
	}

	for i := 0; i < size; i++ {
		pool.next[i] = i + 1
	}
	return pool
}

func (this *Pool) Full() bool {
	this.Lock()
	defer this.Unlock()
	return this.head == this.size
}

func (this *Pool) Get() interface{} {
	// 取出一个节点地址
	this.Lock()
	if this.head == this.size {
		// 申请内存已耗尽 动态扩展TODO
		this.Unlock()
		return reflect.New(this.typ).Interface()
	}
	free := this.head
	nextFree := this.next[free]
	this.head = nextFree
	this.Unlock()

	// 保存节点地址
	obj := this.data.Index(free).Addr().Interface()
	this.addrIndex[obj] = free
	return obj
}

func (this *Pool) Put(obj interface{}) {
	index, ok := this.addrIndex[obj]
	if ok == false {
		return
	}
	this.Lock()
	this.next[index] = this.head
	this.head = index
	this.Unlock()

	return
}

type Pools struct {
	obj   interface{}
	pools []*Pool
}

func NewPools(obj interface{}) *Pools {
	return &Pools{
		obj:   obj,
		pools: []*Pool{},
	}

}
func (this *Pools) Get() interface{} {
	var obj interface{} = nil

	for _, pool := range this.pools {
		if pool.Full() {
			continue
		}
		obj = pool.Get()
	}
	if obj == nil {
		this.pools = append(this.pools, NewPool(this.obj, DEFAULT_OBJECT_SIZE))
		obj = this.Get()
	}
	return obj
}

func (this *Pools) Put(obj interface{}) {
	for _, pool := range this.pools {
		pool.Put(obj)
	}
}
