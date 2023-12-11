package internal

import "sync"

//	type SchedulerInfo struct {
//		NodeName string
//		Address  string
//	}
type SchedulerView struct {
	items []*SchedulerInfo
	lock  *sync.Mutex
	index int
}

func NewSchedulerView() *SchedulerView {
	return &SchedulerView{
		items: make([]*SchedulerInfo, 0),
		lock:  &sync.Mutex{},
		index: 0,
	}
}

func (sv *SchedulerView) Add(s *SchedulerInfo) {
	sv.lock.Lock()
	defer sv.lock.Unlock()
	sv.items = append(sv.items, s)
}

func (sv *SchedulerView) Delete(s *SchedulerInfo) {
	sv.lock.Lock()
	defer sv.lock.Unlock()

	var result []*SchedulerInfo
	for _, info := range sv.items {
		if info.Address != s.Address {
			result = append(result, info)
		}
	}
	sv.items = result
}
func (sv *SchedulerView) GetSchedulerAddr() string {
	sv.lock.Lock()
	defer sv.lock.Unlock()

	sv.index = (sv.index + 1) % len(sv.items)
	return sv.items[sv.index].Address
}
