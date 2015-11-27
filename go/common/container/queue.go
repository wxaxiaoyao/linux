package container

const (
	QUEUE_PRIORITY0 = iota
	QUEUE_PRIORITY1
	QUEUE_PRIORITY2
	QUEUE_PRIORITY3
	QUEUE_PRIORITY4
	QUEUE_PRIORITY5
	QUEUE_PRIORITY6
	QUEUE_PRIORITY_MAX
	QUEUE_PRIORITY_LOW    = QUEUE_PRIORITY5
	QUEUE_PRIORITY_NORMAL = QUEUE_PRIORITY3
	QUEUE_PRIORITY_HIGH   = QUEUE_PRIORITY1
)

type AdaptorInterface interface {
	PushFront(v interface{})
	PushBack(v interface{})
	PopFront() interface{}
	PopBack() interface{}
	Len() int
}
type QueueT struct {
	Queue [QUEUE_PRIORITY_MAX]AdaptorInterface
}

func NewQueue() *QueueT {
	queue := &QueueT{}
	for i := 0; i < QUEUE_PRIORITY_MAX; i++ {
		queue.Queue[i] = NewStaticList()
	}
	return queue
}

func (this *QueueT) NormalIn(v interface{}) {
	this.in(QUEUE_PRIORITY_NORMAL, v)
}

func (this *QueueT) LowIn(v interface{}) {
	this.in(QUEUE_PRIORITY_LOW, v)
}

func (this *QueueT) HighIn(v interface{}) {
	this.in(QUEUE_PRIORITY_HIGH, v)
}

func (this *QueueT) PriorityIn(priority int, v interface{}) {
	this.in(priority, v)
}

func (this *QueueT) Out() interface{} {
	for i := 0; i < QUEUE_PRIORITY_MAX; i++ {
		if v := this.Queue[i].PopFront(); v != nil {
			return v
		}
	}
	return nil
}

func (this *QueueT) NormalOut() interface{} {
	return this.out(QUEUE_PRIORITY_NORMAL)
}

func (this *QueueT) LowOut() interface{} {
	return this.out(QUEUE_PRIORITY_LOW)
}

func (this *QueueT) HighOut() interface{} {
	return this.out(QUEUE_PRIORITY_HIGH)
}

func (this *QueueT) PriorityOut(priority int) interface{} {
	return this.out(priority)
}

func (this *QueueT) out(priority int) interface{} {
	if priority < QUEUE_PRIORITY0 {
		priority = QUEUE_PRIORITY0
	}

	if priority > QUEUE_PRIORITY6 {
		priority = QUEUE_PRIORITY6
	}

	return this.Queue[priority].PopFront()
}

func (this *QueueT) in(priority int, v interface{}) {
	if priority < QUEUE_PRIORITY0 {
		priority = QUEUE_PRIORITY0
	}

	if priority > QUEUE_PRIORITY6 {
		priority = QUEUE_PRIORITY6
	}

	this.Queue[priority].PushBack(v)
}

func (this *QueueT) Len() int {
	size := 0
	for i := 0; i < QUEUE_PRIORITY_MAX; i++ {
		size += this.Queue[i].Len()
	}
	return size
}
