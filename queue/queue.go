package queue

import (
	"sync/atomic"
	"unsafe"
)

type StampedReference struct {
	idx int
	stamp int
}

type DEQueue struct {
	start int
	end int
	bottom int32
	top unsafe.Pointer
}

func NewDEQueue(start int, end int) *DEQueue {
	top := StampedReference{idx: start, stamp: 0}
	queue := DEQueue{top: unsafe.Pointer(&top), bottom: int32(end), start: start, end: end}
	return &queue
}

func loadPointer(p *unsafe.Pointer) *StampedReference {
	if p == nil {
		return nil
	}
	return (*StampedReference)(atomic.LoadPointer(p))
}

func compareAndSwap(p *unsafe.Pointer, old *StampedReference, new *StampedReference) bool {
	var oldUnsafe unsafe.Pointer
	oldUnsafe = nil
	if old != nil {
		oldUnsafe = unsafe.Pointer(old)
	}
	return atomic.CompareAndSwapPointer(p, oldUnsafe, unsafe.Pointer(new))
}

func (dq *DEQueue) PopTop() int {
	oldTop := loadPointer(&dq.top)
	newTop := StampedReference{idx: oldTop.idx + 1, stamp: oldTop.stamp + 1}
	
	if int(dq.bottom) <= oldTop.idx {
		return -1
	}

	if compareAndSwap(&dq.top, oldTop, &newTop) {
		return oldTop.idx
	}

	return -1
}

func (dq *DEQueue) PopBottom() int {
	bottom := int(dq.bottom)
	if bottom == dq.start {
		return -1
	}

	atomic.AddInt32(&dq.bottom, -1)
	bottom--
	particleIdx := bottom
	oldTop := loadPointer(&dq.top)
	newTop := StampedReference{idx: dq.end - 1, stamp: oldTop.stamp + 1}

	if bottom > oldTop.idx {
		return particleIdx
	} 
	if bottom == oldTop.idx {
		atomic.StoreInt32(&dq.bottom, int32(dq.start))
		if compareAndSwap(&dq.top, oldTop, &newTop) {
			return particleIdx
		}
	}
	atomic.StorePointer(&dq.top, unsafe.Pointer(&newTop))
	
	return -1
}