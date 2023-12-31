package server

import (
	"context"
	assignor "coordinator_rpc/RendezousHashing"
	pb "coordinator_rpc/proto"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

var (
	rh = assignor.NewRendezvousHashing()
)

type CoordiantorServer struct{}

func (c CoordiantorServer) AddNodeInfo(ctx context.Context, update *pb.NodeInfoUpdate) (*pb.CoordinatorReply, error) {
	node := &assignor.NodeResource{
		NodeName: update.NodeName,
		HaveCpu:  update.HaveCpu,
		HaveMem:  update.HaveMem,
		Address:  update.Address,
		Port:     "16446",
		Hash:     0,
	}
	actions := rh.AddNode(node)
	err := doActions(actions)
	if err != nil {
		return &pb.CoordinatorReply{
			State:   1,
			Message: fmt.Sprintf("Err:%s", err),
		}, nil
	}
	return &pb.CoordinatorReply{
		State:   0,
		Message: "Success",
	}, nil
}

func (c CoordiantorServer) AddSchedulerInfo(ctx context.Context, update *pb.SchedulerInfoUpdate) (*pb.CoordinatorReply, error) {
	sch := assignor.NewScheduler(update.SchedulerName, update.Address)
	actions := rh.AddScheduler(sch)

	// inform the nodes
	err := doActions(actions)
	if err != nil {
		return &pb.CoordinatorReply{
			State:   1,
			Message: fmt.Sprintf("Err:%s", err),
		}, nil
	}
	//inform all dispatchers
	su := &pb.SchedulerViewUpdate{
		List: []*pb.SchedulerInfo{&pb.SchedulerInfo{
			NodeName: update.SchedulerName,
			Address:  update.Address,
		}},
		Action: "ADD",
	}
	for dispatcherAddr, _ := range rh.Dispatchers {
		fmt.Printf("Add scheduler to dispatcher %s \n", dispatcherAddr)
		conn, _ := grpc.Dial(fmt.Sprintf("%s:16444", dispatcherAddr), grpc.WithInsecure())
		defer conn.Close()
		client := pb.NewDispatcherClient(conn)
		resp, _ := client.UpdateSchedulerView(context.Background(), su)
		log.Printf("client.UpadateNodeResource resp: %d", resp.State)
	}
	psu := &pb.PeerSchedulersUpdate{
		List: []*pb.PeerSchedulerInfo{&pb.PeerSchedulerInfo{
			NodeName: update.SchedulerName,
			Address:  update.Address,
		}},
		Action: "ADD",
	}
	//inform old peer Schedulers
	for _, s := range rh.Schedulers {

		if s.Addr != update.Address {
			fmt.Printf("Add scheduler to peer scheduler %s\n", s.Addr)
			conn, _ := grpc.Dial(fmt.Sprintf("%s:16445", s.Addr), grpc.WithInsecure())
			defer conn.Close()
			client := pb.NewSchedulerClient(conn)
			resp, _ := client.PeerSchedulerUpdate(context.Background(), psu)
			log.Printf("client.UpadateNodeResource resp: %d", resp.State)
		}
	}
	//inform new Scheduler with the old scheduler
	opsu := &pb.PeerSchedulersUpdate{
		List:   nil,
		Action: "ADD",
	}
	opsuList := []*pb.PeerSchedulerInfo{}
	for schedulerAddr, scheduler := range rh.Schedulers {
		if schedulerAddr != update.Address {
			opsuList = append(opsuList, &pb.PeerSchedulerInfo{
				NodeName: scheduler.Name,
				Address:  schedulerAddr,
			})
		}
	}
	fmt.Printf("Add old peer scheduler to old scheduler\n")
	opsu.List = opsuList
	conn, _ := grpc.Dial(fmt.Sprintf("%s:16445", update.Address), grpc.WithInsecure())
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)
	resp, _ := client.PeerSchedulerUpdate(context.Background(), opsu)
	log.Printf("client.UpadateNodeResource resp: %d", resp.State)
	return &pb.CoordinatorReply{
		State:   0,
		Message: "Success",
	}, nil
}

func (c CoordiantorServer) AddDispatcherInfo(ctx context.Context, update *pb.DispatcherInfoUpdate) (*pb.CoordinatorReply, error) {
	fmt.Printf("Add the Dispatcher %s\n", update.SchedulerName)
	disp := &assignor.Dispatcher{
		Name: update.SchedulerName,
		Addr: update.Address,
		Hash: 0,
	}
	schedulers := rh.AddDispatcher(disp)
	transInfo := &pb.SchedulerViewUpdate{
		List:   make([]*pb.SchedulerInfo, 0),
		Action: "ADD",
	}
	for _, s := range schedulers {
		schedulerInfo := &pb.SchedulerInfo{
			NodeName: s.Name,
			Address:  s.Addr,
		}
		transInfo.List = append(transInfo.List, schedulerInfo)
	}
	connDispatcher, _ := grpc.Dial(fmt.Sprintf("%s:16444", update.Address), grpc.WithInsecure())
	defer connDispatcher.Close()
	client := pb.NewDispatcherClient(connDispatcher)

	resp, _ := client.UpdateSchedulerView(context.Background(), transInfo)
	log.Printf("client.UpadateNodeResource resp: %d", resp.State)

	return &pb.CoordinatorReply{
		State:   0,
		Message: "Success",
	}, nil
}
func doActions(actions []*assignor.TransInfo) error {

	mergeActions := mergeTransInfos(actions)
	var wg sync.WaitGroup

	startTime := time.Now()
	for _, action := range mergeActions {
		wg.Add(1)

		go func(action *assignor.TransInfo) {
			defer wg.Done()
			if action.Action == "ADD" {
				fmt.Printf("Add the %d nodes to %s\n", len(action.NodeResourceList), rh.GetSchedulerNameByAddr(action.SourceAddr))
			} else {
				fmt.Printf("Move the %d nodes from %s to %s\n", len(action.NodeResourceList), rh.GetSchedulerNameByAddr(action.SourceAddr), rh.GetSchedulerNameByAddr(action.TargetAddr))
			}
			nru := TransInfo2NodeUpdate(action)
			conn, err := grpc.Dial(fmt.Sprintf("%s:16445", action.SourceAddr), grpc.WithInsecure())
			if err != nil {
				fmt.Println("Error connecting to gRPC server:", err)
				return
			}
			defer conn.Close()

			client := pb.NewSchedulerClient(conn)

			_, err = client.UpadateNodeResource(context.Background(), nru)
			if err != nil {
				fmt.Println("Error calling gRPC UpadateNodeResource:", err)
				return
			}

			//fmt.Printf("%d\n", resp.State)
		}(action)
	}

	wg.Wait()
	fmt.Println("All goroutines have finished.")
	// 输出执行时间
	endTime := time.Now()
	executionTime := endTime.Sub(startTime)
	fmt.Println("Execution Time:", executionTime)
	rh.Statisics()
	return nil

	//for _, action := range mergeActions {
	//	if action.Action == "ADD" {
	//		fmt.Printf("Add the %d nodes to %s\n", len(action.NodeResourceList), rh.GetSchedulerNameByAddr(action.SourceAddr))
	//	} else {
	//		fmt.Printf("Move the %d nodes from %s to %s\n", len(action.NodeResourceList), rh.GetSchedulerNameByAddr(action.SourceAddr), rh.GetSchedulerNameByAddr(action.TargetAddr))
	//	}
	//	nru := TransInfo2NodeUpdate(action)
	//	conn, _ := grpc.Dial(fmt.Sprintf("%s:16445", action.SourceAddr), grpc.WithInsecure())
	//	defer conn.Close()
	//	client := pb.NewSchedulerClient(conn)
	//	startTime := time.Now()
	//	resp, _ := client.UpadateNodeResource(context.Background(), nru)
	//	endTime := time.Now()
	//	executionTime := endTime.Sub(startTime)
	//	cost += executionTime
	//	fmt.Printf("%d\n", resp.State)
	//}

}
func TransInfo2NodeUpdate(action *assignor.TransInfo) *pb.NodeResourceUpdate {
	//type TransInfo struct {
	//	Action           string
	//	NodeResourceList []*NodeResource
	//	SourceAddr       string
	//	TargetAddr       string
	//}
	nru := &pb.NodeResourceUpdate{
		List:       nil,
		Action:     action.Action,
		SourceAddr: action.SourceAddr,
		TargetAddr: action.TargetAddr,
	}
	list := make([]*pb.NodeResource, 0)
	for _, nodeResource := range action.NodeResourceList {
		list = append(list, &pb.NodeResource{
			NodeName: nodeResource.NodeName,
			HaveCpu:  nodeResource.HaveCpu,
			HaveMem:  nodeResource.HaveMem,
			Address:  nodeResource.Address,
		})
	}
	nru.List = list
	return nru
	//type NodeResourceUpdate struct {
	//	state         protoimpl.MessageState
	//	sizeCache     protoimpl.SizeCache
	//	unknownFields protoimpl.UnknownFields
	//
	//	List       []*NodeResource `protobuf:"bytes,1,rep,name=list,proto3" json:"list,omitempty"`
	//	Action     string          `protobuf:"bytes,2,opt,name=Action,proto3" json:"Action,omitempty"`
	//	SourceAddr string          `protobuf:"bytes,3,opt,name=SourceAddr,proto3" json:"SourceAddr,omitempty"`
	//	TargetAddr string          `protobuf:"bytes,4,opt,name=TargetAddr,proto3" json:"TargetAddr,omitempty"`
	//}

}
func mergeTransInfos(infos []*assignor.TransInfo) []*assignor.TransInfo {
	transInfos := make(map[string]*assignor.TransInfo)

	for _, info := range infos {
		if info.Action == "ADD" {
			if transInfo, ok := transInfos[info.SourceAddr]; ok {
				transInfo.NodeResourceList = append(transInfo.NodeResourceList, info.NodeResourceList...)
			} else {
				transInfos[info.SourceAddr] = info
			}
		} else { // DEL
			key := "DEL" + info.SourceAddr + "-" + info.TargetAddr
			if transInfo, ok := transInfos[key]; ok {
				transInfo.NodeResourceList = append(transInfo.NodeResourceList, info.NodeResourceList...)
			} else {
				transInfos[key] = info
			}
		}
	}
	actions := make([]*assignor.TransInfo, 0)
	for _, v := range transInfos {
		actions = append(actions, v)
	}
	return actions
}
