package main

import (
	"MinusFifteen/pb"
	"fmt"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
)

var globalNodeID atomic.Int32
var TotalStorageBytes atomic.Int64

func main(){
	fmt.Println("Starting Server...")

	lis, err := net.Listen("tcp",":50051")
	if err != nil{
		fmt.Println("Failed to Listen:",err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterCollectorServer(s, &Server{})

	go WebServer()

	fmt.Println("Starting 3 nodes...")
	startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)
    startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)
    startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)

	fmt.Println("Server listening on :50051")
	if err := s.Serve(lis); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}