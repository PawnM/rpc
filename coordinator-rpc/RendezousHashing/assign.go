package RendezousHashing

import (
	"fmt"
	"hash/fnv"
)

type UpdateInfo struct {
	Action           string
	NodeResourceList []*NodeResource
	SchedulerAddr    string
}

// HashFunc represents the hash function type
type HashFunc func(data string) uint32

type NodeResource struct {
	NodeName string
	HaveCpu  int64
	HaveMem  int64
	Address  string
	Port     string
	Hash     uint32
}

// NewNode creates a new node with a given ID
func NewNode(name string) *NodeResource {
	return &NodeResource{
		NodeName: name,
		HaveMem:  0,
		HaveCpu:  0,
		Address:  "127.0.0.1",
		Port:     "202",
	}
}

type Scheduler struct {
	Name string
	Addr string
	Hash uint32
}

func NewScheduler(name string, addr string) *Scheduler {
	return &Scheduler{Name: name, Addr: addr}
}

// RendezvousHashing implements the Rendezvous Hashing algorithm
type RendezvousHashing struct {
	Nodes      map[string]*NodeResource
	Schedulers map[string]*Scheduler
	SNView     map[string]string // node -> scheduler
	Hash       HashFunc
}

func NewRendezvousHashing() *RendezvousHashing {
	return &RendezvousHashing{
		Nodes:      make(map[string]*NodeResource),
		Schedulers: make(map[string]*Scheduler),
		SNView:     make(map[string]string),
		Hash:       fnvHash,
	}
}
func (rh *RendezvousHashing) Statisics() {
	statis := make(map[string]int)
	for _, v := range rh.SNView {
		statis[v] += 1
	}
	for k, v := range statis {
		fmt.Printf("%s:%d\n", k, v)
	}
	minValue, maxValue := findMinMaxValues(statis)
	fmt.Printf("最小值：%d\n", minValue)
	fmt.Printf("最大值：%d\n", maxValue)
	fmt.Printf("==================================\n")
}
func findMinMaxValues(myMap map[string]int) (int, int) {
	var minValue, maxValue int

	// 初始化最小值和最大值
	for _, value := range myMap {
		minValue = value
		maxValue = value
		break
	}

	// 遍历map，更新最小值和最大值
	for _, value := range myMap {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	return minValue, maxValue
}

// AddNode adds a node to the RendezvousHashing instance
func (rh *RendezvousHashing) AddNode(node *NodeResource) []*UpdateInfo {

	var selectedScheduler *Scheduler
	maxHash := uint32(0)

	for _, s := range rh.Schedulers {
		hash := rh.Hash(s.Name + node.NodeName)
		if hash > maxHash {
			maxHash = hash
			selectedScheduler = s
		}
	}
	node.Hash = fnvHash(node.NodeName)
	rh.Nodes[node.NodeName] = node
	// inform the scheduler to add the node
	rh.SNView[node.NodeName] = selectedScheduler.Name
	actions := make([]*UpdateInfo, 1)
	actions[0] = &UpdateInfo{
		Action:           "ADD",
		NodeResourceList: []*NodeResource{node},
		SchedulerAddr:    selectedScheduler.Addr,
	}
	return actions
}
func (rh *RendezvousHashing) DeleteNode(node *NodeResource) []*UpdateInfo {
	sname := rh.SNView[node.NodeName]
	s := rh.Schedulers[sname]
	// 获取原scheduler
	delete(rh.SNView, node.NodeName)
	delete(rh.Nodes, node.NodeName)
	// inform the scheduler to delete the node
	return []*UpdateInfo{
		&UpdateInfo{
			Action:           "DEL",
			NodeResourceList: []*NodeResource{node},
			SchedulerAddr:    s.Addr,
		},
	}
}

func (rh *RendezvousHashing) AddScheduler(scheduler *Scheduler) []*UpdateInfo {
	actions := make([]*UpdateInfo, 0)
	rh.Schedulers[scheduler.Name] = scheduler
	for nodeName, nodeResource := range rh.Nodes {
		hash := rh.Hash(scheduler.Name + nodeName)

		//fmt.Printf("%d - %d\n", hash, nodeResource.Hash)
		if hash > nodeResource.Hash {
			oldScheduler := rh.Schedulers[rh.SNView[nodeName]]
			// remove the node from the schueduler
			actions = append(actions, &UpdateInfo{
				Action:           "DEL",
				NodeResourceList: []*NodeResource{nodeResource},
				SchedulerAddr:    oldScheduler.Addr,
			})
			nodeResource.Hash = hash
			actions = append(actions, &UpdateInfo{
				Action:           "ADD",
				NodeResourceList: []*NodeResource{nodeResource},
				SchedulerAddr:    scheduler.Addr,
			})
			rh.SNView[nodeName] = scheduler.Name
		}
	}
	return actions
}
func (rh *RendezvousHashing) DeleteScheduler(scheduler *Scheduler) []*UpdateInfo {
	actions := make([]*UpdateInfo, 0)

	reBalanceNode := make([]*NodeResource, 0)
	for nname, sname := range rh.SNView {
		if sname == scheduler.Name {
			nr := rh.Nodes[nname]
			reBalanceNode = append(reBalanceNode, nr)
		}
	}

	for _, nodeResource := range reBalanceNode {
		actions = append(actions, rh.DeleteNode(nodeResource)...)
	}
	delete(rh.Schedulers, scheduler.Name)
	for _, nodeResource := range reBalanceNode {
		actions = append(actions, rh.AddNode(nodeResource)...)
	}

	return actions
}

// fnvHash is a simple hash function using FNV-1a algorithm
func fnvHash(data string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(data))
	return h.Sum32()
}
