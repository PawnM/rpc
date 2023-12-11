package reblance

type NodeResource struct {
	NodeName string
	HaveCpu  int64
	HaveMem  int64
	Address  string
	Port     string
}

type UpdateInfo struct {
	Action           string
	NodeResourceList []*NodeResource
	SchedulerAddr    string
}

type StickyAssignor struct {
	Dispatcher_map  map[string]string
	Scheduler_map   map[string]string
	Virtualnode_map map[string]*NodeResource
	partitionMap    map[string][]string // scheduler -> nodes
}

//
//func NewStickyAssignor() *StickyAssignor {
//	return &StickyAssignor{
//		Dispatcher_map:  make(map[string]string),
//		Scheduler_map:   make(map[string]string),
//		Virtualnode_map: make(map[string]*NodeResource, 0),
//		partitionMap:    make(map[string][]string),
//	}
//}
//func (s *StickyAssignor) StickyAssign(action string, func_type string) []*UpdateInfo {
//	// 如果老的视角为空，则roubin分配
//
//	//
//
//}
//func (s *StickyAssignor) RoundRobinAssignor() []*UpdateInfo {
//	//type UpdateInfo struct {
//	//	Action           string
//	//	NodeResourceList []*NodeResource
//	//	SchedulerAddr    string
//	//}
//	actions := make([]*UpdateInfo, len(s.Scheduler_map))
//	i := 0
//	for _, v := range s.Scheduler_map {
//		actions[i] = &UpdateInfo{
//			Action:           "ADD",
//			NodeResourceList: make([]*NodeResource, 0),
//			SchedulerAddr:    v,
//		}
//		i++
//	}
//
//	for k, v := range s.Virtualnode_map {
//		actions[i%len(s.Scheduler_map)].NodeResourceList = append(actions[i%len(s.Scheduler_map)].NodeResourceList)
//	}
//}
