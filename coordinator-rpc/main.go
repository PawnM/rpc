package main

import (
	rdh "coordinator_rpc/RendezousHashing"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// Create a RendezvousHashing instance
	rh := rdh.NewRendezvousHashing()
	scheduler1 := rdh.NewScheduler("Scheduler1", "192.168.1.1")
	printInfo(rh.AddScheduler(scheduler1))
	// Add nodes to the system
	for i := 1; i <= 10000; i++ {
		time.Sleep(10)
		rh.AddNode(generateNode())
	}
	schedulerGenerator := generateSchedulerFunc()
	printInfo(rh.AddScheduler(schedulerGenerator()))
	rh.Statisics()
	printInfo(rh.AddScheduler(schedulerGenerator()))
	rh.Statisics()
	printInfo(rh.AddScheduler(schedulerGenerator()))
	rh.Statisics()
	printInfo(rh.AddScheduler(schedulerGenerator()))
	rh.Statisics()
	printInfo(rh.AddScheduler(schedulerGenerator()))
	rh.Statisics()
	//printInfo(rh.DeleteScheduler(scheduler2))
}
func printInfo(infos []*rdh.UpdateInfo) {
	newInfos := mergeInfos(infos)
	addNum, delNum := 0, 0
	addNodeNum, delNodeNum := 0, 0
	for _, up := range newInfos {
		if up.Action == "ADD" {
			addNum += 1
			addNodeNum += len(up.NodeResourceList)
			fmt.Printf("add %d node to the %s\n", len(up.NodeResourceList), up.SchedulerAddr)
		} else {
			delNum += 1
			delNodeNum += len(up.NodeResourceList)
			fmt.Printf("del %d node from the %s\n", len(up.NodeResourceList), up.SchedulerAddr)
		}
	}
	fmt.Println("------------------------------------")
}
func mergeInfos(infos []*rdh.UpdateInfo) []*rdh.UpdateInfo {
	addInfos := make(map[string]*rdh.UpdateInfo)
	delInfos := make(map[string]*rdh.UpdateInfo)

	for _, info := range infos {
		if info.Action == "ADD" {
			if addInfo, ok := addInfos[info.SchedulerAddr]; ok {
				addInfo.NodeResourceList = append(addInfo.NodeResourceList, info.NodeResourceList...)
			} else {
				addInfos[info.SchedulerAddr] = info
			}
		} else {
			if delInfo, ok := delInfos[info.SchedulerAddr]; ok {
				delInfo.NodeResourceList = append(delInfo.NodeResourceList, info.NodeResourceList...)
			} else {
				delInfos[info.SchedulerAddr] = info
			}
		}
	}
	actions := make([]*rdh.UpdateInfo, 0)
	for _, v := range addInfos {
		actions = append(actions, v)
	}
	for _, v := range delInfos {
		actions = append(actions, v)
	}
	return actions
}
func generateSchedulerFunc() func() *rdh.Scheduler {
	id := 0

	return func() *rdh.Scheduler {
		l := randomString(6)
		id += 1
		return rdh.NewScheduler(fmt.Sprintf("Scheduler-%s", l), fmt.Sprintf("192.168.1.%d", id))
	}
}
func generateNode() *rdh.NodeResource {
	l := randomString(6)
	fmt.Printf("Add node %s\n", fmt.Sprintf("Node-%s", l))
	return rdh.NewNode(fmt.Sprintf("Node-%s", l))
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
