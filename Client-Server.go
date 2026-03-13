package main

import (
	"MinusFifteen/pb"
	"context"
	"fmt"
	"math/rand"
	"time"
	"sync"
	"sync/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	activeNodes = make(map[int]context.CancelFunc)
	activeNodesMu sync.Mutex
)

func randomUniform(min, max float32)float32{
	return min + rand.Float32()*(max-min)
}

type NodeInfo struct{
	NodeNumber int
}

type NodeState struct {
	NodeID    int     `json:"node_id"`
	Temp      float32 `json:"temp"`
	Humidity  float32 `json:"humidity"`
	HeatIndex float32 `json:"heat_index"`
}

var (
	latestData = make(map[int]NodeState)
	dataMu     sync.Mutex
)

type Server struct{
	nodes [] any
	mu sync.Mutex
	currentSize uint64 //56 Bytes per json
	pb.UnimplementedCollectorServer
}

func (s *Server) ReportStatus(stream pb.Collector_ReportStatusServer) error{
	for{
		msg, err := stream.Recv()
		if err != nil{
			return err
		}
		atomic.AddUint64(&s.currentSize, 56)
		TotalStorageBytes.Add(56)
		dataMu.Lock()
        latestData[int(msg.NodeNumber)] = NodeState{
            NodeID:    int(msg.NodeNumber),
            Temp:      msg.Temperature_C,
            Humidity:  msg.Humidity,
            HeatIndex: msg.HeatIndex,
        }
        dataMu.Unlock()
		fmt.Printf("Recieved Data: {Node Number: %v\n Temp_C: %v\nHumidity: %v\nHeatIndex: %v}\n", msg.NodeNumber, msg.Temperature_C, msg.Humidity, msg.HeatIndex)
		fmt.Printf("Server Storage Size: %v\n\n", s.currentSize)
	}
}


type Node struct{
	NodeNumber int
	killChannel chan any
}

func startNode(node *Node, MessagesPerSecond int){
	ctx, cancel := context.WithCancel(context.Background())
	id := node.NodeNumber

	activeNodesMu.Lock()
	activeNodes[id] = cancel
	activeNodesMu.Unlock()

	go func(){
		conn,_ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		client := pb.NewCollectorClient(conn)
		stream, _ := client.ReportStatus(context.Background())
		

		for{
			select{
			
			case <- ctx.Done():
				fmt.Printf("Node %d closing...\n\n", id)
				stream.CloseSend()
				conn.Close()
				return
			
			default: 
				message := &pb.DataPoint{
				NodeNumber: float32(id),
				Temperature_C: randomUniform(20,30),
				Humidity: randomUniform(20,30),
				HeatIndex: randomUniform(20,30),
			}	
			stream.Send(message)
			time.Sleep(time.Duration(1000/MessagesPerSecond) * time.Millisecond)
			}
		}
	}()
}