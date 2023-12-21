package main

import (
	pb "dispatcher_rpc/proto"
	"dispatcher_rpc/server"
	"flag"
	"google.golang.org/grpc"
	"net"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "p", "16444", "启动端口号")
	flag.Parse()
}

func main() {
	//logFile, err := os.OpenFile("./logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatal("Error opening log file:", err)
	//}
	//defer logFile.Close()
	//log.SetOutput(logFile)

	s := grpc.NewServer()
	pb.RegisterDispatcherServer(s, &server.DispatcherServer{})
	lis, _ := net.Listen("tcp", ":"+port)
	s.Serve(lis)
}
