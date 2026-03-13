package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	activeNodes   = make(map[int]context.CancelFunc)
	activeNodesMu sync.Mutex
	latestData    = make(map[int]NodeState)
	dataMu        sync.Mutex
)

type NodeState struct {
	NodeID    int     `json:"node_id"`
	Temp      float32 `json:"temp"`
	Humidity  float32 `json:"humidity"`
	HeatIndex float32 `json:"heat_index"`
}

type Node struct {
	NodeNumber int
}

func randomUniform(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func startNode(node *Node, MessagesPerSecond int) {
	ctx, cancel := context.WithCancel(context.Background())
	id := node.NodeNumber

	activeNodesMu.Lock()
	activeNodes[id] = cancel
	activeNodesMu.Unlock()

	go func() {
		opts := mqtt.NewClientOptions().
			AddBroker("tcp://localhost:1883").
			SetClientID(fmt.Sprintf("Node-%d-Pub", id))
            
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("Node %d failed to connect to broker.\n", id)
			return
		}

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Node %d disconnecting from broker...\n", id)
				client.Disconnect(250) 
				return

			default:
				payload := NodeState{
					NodeID:    id,
					Temp:      randomUniform(20, 30),
					Humidity:  randomUniform(20, 30),
					HeatIndex: randomUniform(20, 30),
				}
				
				bytes, _ := json.Marshal(payload)
				client.Publish("minusfifteen/data", 0, false, bytes)

				time.Sleep(time.Duration(1000/MessagesPerSecond) * time.Millisecond)
			}
		}
	}()
}