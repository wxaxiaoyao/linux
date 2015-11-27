package pool

import (
	"sync"
)

var (
	DEFAULT_BYTE_POOL_SIZE = 10000000
	IndexArray             = []int{8, 16, 32, 64, 128, 256, 1024, 2048}
	IndexNum               = len(IndexArray)
)

type ByteSlice struct {
	start int
	end   int
	Data  []byte
}
type BytePool struct {
	*sync.Mutex
	data          []byte
	start         int
	end           int
	size          int
	usedList      []*StaticList
	freeList      []*StaticList
	byteSlicePool *Pools
}

func NewBytePool() *BytePool {
	bytePool := &BytePool{
		Mutex:         &sync.Mutex{},
		data:          make([]byte, DEFAULT_BYTE_POOL_SIZE),
		size:          DEFAULT_BYTE_POOL_SIZE,
		end:           DEFAULT_BYTE_POOL_SIZE,
		usedList:      []*StaticList{},
		freeList:      []*StaticList{},
		byteSlicePool: NewPools(BytePool{}),
	}
	for i := 0; i < IndexNum; i++ {
		bytePool.usedList[i] = NewStaticList()
		bytePool.freeList[i] = NewStaticList()
	}
	return bytePool
}

func (this *BytePool) Index(size int) int {
	for i := 0; i < IndexNum; i++ {
		if size <= IndexArray[i] {
			return i
		}
	}
	return IndexNum
}

func (this *BytePool) Get(size int) *ByteSlice {
	index := this.Index(size)

	// 申请大块内存 size > 2048
	if index == IndexNum {
		bs := this.byteSlicePool.Get().(*ByteSlice)
		bs.end = this.end
		bs.start = this.end - size
		this.end -= bs.start
		bs.Data = this.data[bs.start:bs.end]
		return bs
	}
	bs := this.freeList[index].PopFront()
	if bs == nil {

	}
}
