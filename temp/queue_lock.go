package queue

import (
	"sync/atomic"
	"unsafe"
	"sync"
	// "fmt"
)

type StampedReference struct {
	idx int
	stamp int
}

// type Bottom struct {
// 	idx int
// 	mutex *sync.Mutex
// }

type DEQueue struct {
	start int
	bottom int
	top *StampedReference
	mutex *sync.Mutex
}

func NewDEQueue(start int, end int) *DEQueue {
	top := StampedReference{idx: start, stamp: 0}
	var mutex sync.Mutex
	// bottom := Bottom{idx: end, mutex: &mutex}
	queue := DEQueue{mutex: &mutex, top: &top, bottom: end, start: start}
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
	dq.mutex.Lock()
	oldTop := dq.top
	newTop := StampedReference{idx: oldTop.idx + 1, stamp: oldTop.stamp + 1}
	
	// dq.bottom.mutex.Lock()
	if dq.bottom <=	oldTop.idx {
		// dq.bottom.mutex.Unlock()
		dq.mutex.Unlock()
		return -1
	}
	// dq.bottom.mutex.Unlock()
	dq.top = &newTop
	dq.mutex.Unlock()
	return oldTop.idx
	// particleIdx := oldTop.idx
	// if compareAndSwap(&dq.top, oldTop, &newTop) {
	// 	return particleIdx
	// }

	// return -1
}

func (dq *DEQueue) PopBottom() int {
	// bottom := int(dq.bottom)
	dq.mutex.Lock()
	if dq.bottom == dq.start {
		dq.mutex.Unlock()
		return -1
	}
	
	dq.bottom--
	// atomic.AddInt32(&dq.bottom, -1)
	// bottom--
	// dq.bottom.mutex.Lock()
	// dq.bottom.idx--
	// particleIdx := bottom
	oldTop := dq.top
	newTop := StampedReference{idx: dq.start	, stamp: oldTop.stamp + 1}

	// if oldTop == loadPointer(&dq.top) {	//ISSUE IS TOP
	if dq.bottom > oldTop.idx {
		// dq.bottom.mutex.Unlock()
		dq.mutex.Unlock()
		return dq.bottom
	} 
	if dq.bottom == oldTop.idx {
		// atomic.StoreInt32(&dq.bottom, int32(dq.start))
		dq.bottom = dq.start
		dq.top = &newTop
		dq.mutex.Unlock()
		return dq.bottom
		// if compareAndSwap(&dq.top, oldTop, &newTop) {
		// 	// dq.bottom.mutex.Unlock()
		// 	return particleIdx
		// }
	}
	// atomic.StorePointer(&dq.top, unsafe.Pointer(&newTop))
	dq.top = &newTop
	dq.mutex.Unlock()
	// dq.bottom.mutex.Unlock()
	return -1
	// }
	// dq.bottom.mutex.Unlock()
	// return -1
}

func (dq *DEQueue) IsEmpty() bool {
	dq.mutex.Lock()
	currTop := dq.top.idx
	currBottom := dq.bottom
	dq.mutex.Unlock()
	return currTop < currBottom
}