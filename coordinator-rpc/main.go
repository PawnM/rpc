package main

import (
	pb "coordinator_rpc/proto"
	"coordinator_rpc/server"
	"flag"
	"google.golang.org/grpc"
	"net"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "p", "16000", "启动端口号")
	flag.Parse()
}
func main() {
	s := grpc.NewServer()
	pb.RegisterCoordinatorServer(s, &server.CoordiantorServer{})
	lis, _ := net.Listen("tcp", ":"+port)
	s.Serve(lis)
}
