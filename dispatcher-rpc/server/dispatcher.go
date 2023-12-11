package server

import (
	"context"
	"dispatcher_rpc/internal"
	pb "dispatcher_rpc/proto"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	funcView      = internal.NewFuncView() // 支持查找变更 func
	SchedulerView = internal.NewSchedulerView()
	localIp       = getLocalIPv4().String()
	ConnCache     = NewLRUCache(20)
)

type DispatcherServer struct{}

func (d DispatcherServer) Statis(ctx context.Context, request *pb.UserRequest) (*pb.UserRequestReply, error) {
	return &pb.UserRequestReply{
		RequestId:   1,
		FuncName:    "request.FuncName",
		Destination: "result",
	}, nil
}
func (d DispatcherServer) UpdateInstanceView(ctx context.Context, update *pb.InstanceUpdate) (*pb.InstanceUpdateReply, error) {
	list := update.List
	for _, v := range list {
		t := time.Now().UnixNano() / 1e6
		item := &internal.FuncInfo{FuncName: v.FuncName,
			Address:   v.Address,
			Timestamp: t,
			State:     true}
		if update.Action == "ADD" {
			funcView.Add(item)
			log.Printf("Add new instance -%d-%s-%s at -%d\n", v.RequestId, item.FuncName, item.Address, time.Now().UnixNano()/1e6)
			//fmt.Printf("Add new instance -%d-%s-%s at -%d\n", v.RequestId, item.FuncName, item.Address, time.Now().UnixNano()/1e6)
		} else if update.Action == "DELETE" {
			funcView.Delete(item)
		}
	}
	return &pb.InstanceUpdateReply{
		State: 0,
	}, nil
}

func (d DispatcherServer) Dispatch(ctx context.Context, userRequests *pb.UserRequestList) (*pb.UserRequestReply, error) {
	t := time.Now().UnixNano() / 1e6
	for _, request := range userRequests.GetList() {

		result := funcView.Dispatch(request.FuncName)
		if result == "" {
			log.Printf("Need to scale up -%d-%s at -%d\n", request.RequestId, request.FuncName, t)
			//fmt.Printf("Need to scale up -%d-%s at -%d\n", request.RequestId, request.FuncName, time.Now().UnixNano()/1e6)
			// scale
			schedulerAddr := SchedulerView.GetSchedulerAddr()

			conn := ConnCache.Get(schedulerAddr)
			if conn != nil {
				fmt.Printf("Cache hitted!\n")
			} else {
				conn, _ = grpc.Dial(fmt.Sprintf("%s:16445", schedulerAddr), grpc.WithInsecure())
				ConnCache.Put(schedulerAddr, conn)
				fmt.Printf("Cache miss!\n")
			}

			client := pb.NewSchedulerClient(conn)
			_, _ = client.Schedule(context.Background(), &pb.ScheduleRequest{
				RequestId:      request.RequestId,
				FuncName:       request.FuncName,
				RequireCpu:     request.RequireCpu,
				RequireMem:     request.RequireMem,
				DispatcherAddr: localIp,
			})
			//fmt.Printf("Need to scale up %d:%s to %s\n", request.RequestId, request.FuncName, schedulerAddr)
			// choose a scheudler
		} else {
			//fmt.Printf("Route request %d:%s to %s\n", request.RequestId, request.FuncName, result)
		}
	}

	return &pb.UserRequestReply{
		RequestId:   0,
		FuncName:    "0",
		Destination: "0",
	}, nil
}
func (d DispatcherServer) UpdateSchedulerView(ctx context.Context, update *pb.SchedulerViewUpdate) (*pb.SchedulerViewUpdateReply, error) {

	list := update.List
	for _, v := range list {
		item := &internal.SchedulerInfo{NodeName: v.NodeName, Address: v.Address}
		if update.Action == "ADD" {
			fmt.Printf("Add new scheduelr %s:%s\n", item.NodeName, item.Address)
			SchedulerView.Add(item)
		} else if update.Action == "DELETE" {
			SchedulerView.Delete(item)
		}
	}
	return &pb.SchedulerViewUpdateReply{State: 0}, nil
}
func getLocalIPv4() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Println(ipNet.IP.String())
				return ipNet.IP
			}
		}
	}

	return nil
}
