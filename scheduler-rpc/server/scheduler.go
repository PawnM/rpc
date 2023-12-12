package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"scheduler_rpc/internal"
	"scheduler_rpc/internal/cache"
	pb "scheduler_rpc/proto"
	"strings"
	"time"
)

var (
	nodeView                *cache.Cache
	nodeResourceUpdateQueue *cache.FIFO
	RequestQueue            *internal.PriorityQueue
	NotDeployed             map[int64]struct{}
	ConnCache               *LRUCache
)

func init() {
	RequestQueue = internal.NewPriorityQueue()
	nodeResourceUpdateQueue = cache.NewFIFO()
	nodeView = cache.NewCache()
	NotDeployed = make(map[int64]struct{})
	ConnCache = NewLRUCache(20)
}

type SchedulerServer struct{}

func (s SchedulerServer) Schedule(ctx context.Context, requests *pb.ScheduleRequestList) (*pb.ScheduleReply, error) {
	// 根据将请求入列
	priority := time.Now().UnixNano() / 1e6
	for _, request := range requests.List {
		log.Printf("Scheduler receive -%d- at -%d\n", request.RequestId, priority)
		//fmt.Printf("Scheduler receive -%d- at -%d\n", request.RequestId, priority)
		reqestInfo := &internal.RequestInfo{RequestId: request.RequestId, FunctionName: request.FuncName, RequireCpu: request.RequireCpu, RequireMem: request.RequireMem, DispatcherAddr: request.DispatcherAddr}
		if strings.Contains(request.FuncName, "galaxy-") {
			priority = 0
		}
		//fmt.Printf("Receive new %d:%s\n", request.RequestId, request.FuncName)
		RequestQueue.Push(&internal.RequestItem{reqestInfo, priority})
	}

	return &pb.ScheduleReply{
		RequestId:      0,
		FuncName:       "",
		DeployPosition: "",
	}, nil
}

func (s SchedulerServer) UpadateNodeResource(ctx context.Context, update *pb.NodeResourceUpdate) (*pb.NodeResourceReply, error) {
	// 添加到队列中
	action := update.Action
	for _, nodeResource := range update.List {

		fmt.Printf("%s new NodeResource: %s\n", action, nodeResource.NodeName)
		item := &internal.NodeResourceItem{
			Action: action,
			Node: internal.NodeResource{
				NodeName: nodeResource.NodeName,
				HaveCpu:  nodeResource.HaveCpu,
				HaveMem:  nodeResource.HaveMem,
				Address:  nodeResource.Address,
				Port:     nodeResource.Port,
			},
		}
		nodeResourceUpdateQueue.Enqueue(item)
	}
	return &pb.NodeResourceReply{
		State: 0,
	}, nil
}

//	type NodeResourceItem struct {
//		Action string
//		Node   NodeResource
//	}

func ResourceUpdate() {
	for {
		it := nodeResourceUpdateQueue.Dequeue()
		if it == nil {
			continue
		}
		if it.Action == "ADD" {
			nodeView.Set(it.Node.NodeName, &it.Node)
		} else if it.Action == "DELETE" {
			nodeView.Delete(it.Node.NodeName)
		}
	}
}
func Schedule() {
	for {
		// 拿到请求
		i := RequestQueue.Pop()
		it := i.Value
		log.Printf("Scheudler handle request %d\n", it.RequestId)
		hasDeployed := false
		for nodeName, node := range nodeView.Cache {
			if node.HaveCpu >= it.RequireCpu && node.HaveMem >= it.RequireMem {
				log.Printf("deploy -%d-%s to node %s at -%d\n", it.RequestId, it.FunctionName, nodeName, time.Now().UnixNano()/1e6)
				node.HaveMem -= it.RequireMem
				node.HaveCpu -= it.RequireCpu
				nodeView.Set(nodeName, node)
				// deploy to node
				conn := ConnCache.Get(node.Address)
				if conn != nil {
					fmt.Printf("Cache hitted!\n")
				} else {
					conn, _ = grpc.Dial(fmt.Sprintf("%s:16446", node.Address), grpc.WithInsecure())
					ConnCache.Put(node.Address, conn)
					fmt.Printf("Cache miss!\n")
				}
				clinet := pb.NewNodeClient(conn)
				_, _ = clinet.Deploy(context.Background(), &pb.InstanceDeploy{List: []*pb.NodeInstanceInfo{
					{
						RequestId:      it.RequestId,
						FuncName:       it.FunctionName,
						DispatcherAddr: it.DispatcherAddr,
					},
				}})
				hasDeployed = true
				break
			}
		}
		if !hasDeployed {
			fmt.Printf("Scheduler cant find the deployed position for %d:%s.\n", it.RequestId, it.FunctionName)
		}

		//if !hasDeployed {
		//	if _, exists := NotDeployed[it.RequestId]; !exists {
		//		ur := &pb.UserRequest{
		//			RequestId:  it.RequestId,
		//			FuncName:   it.FunctionName,
		//			RequireCpu: it.RequireCpu,
		//			RequireMem: it.RequireMem,
		//		}
		//		connDispatcher, _ := grpc.Dial(it.DispatcherAddr+":16444", grpc.WithInsecure())
		//		defer connDispatcher.Close()
		//		client := pb.NewDispatcherClient(connDispatcher)
		//		_, _ = client.Dispatch(context.Background(), &pb.UserRequestList{
		//			List: []*pb.UserRequest{ur},
		//		})
		//		NotDeployed[it.RequestId] = struct{}{}
		//	} else {
		//		fmt.Printf("Scheduler cant find the deployed position for %d:%s.\n", it.RequestId, it.FunctionName)
		//	}
		//}
	}
}
