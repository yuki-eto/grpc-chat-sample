package dao

import "sync"

type IncrementID struct {
	id  uint64
	mtx *sync.Mutex
}

func NewIncrementID() *IncrementID {
	return &IncrementID{
		id:  0,
		mtx: &sync.Mutex{},
	}
}

func (i *IncrementID) Get() uint64 {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	i.id++
	return i.id
}
