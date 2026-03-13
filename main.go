package main

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var globalNodeID atomic.Int32
var TotalStorageBytes atomic.Uint64

func main(){
	fmt.Println("Starting Server...")

	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("Testbed-Server-12345")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("CRITICAL: Could not connect to public broker.")
		panic(token.Error())
	}

	client.Subscribe("minusfifteen/data", 0, func(client mqtt.Client, msg mqtt.Message) {
		var data NodeState
		json.Unmarshal(msg.Payload(), &data)

		TotalStorageBytes.Add(56) 
		
		dataMu.Lock()
		latestData[data.NodeID] = data
		dataMu.Unlock()
	})

	go WebServer()

	// fmt.Println("Starting 3 nodes...")
	// startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)
    // startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)
    // startNode(&Node{NodeNumber: int(globalNodeID.Add(1))}, 1)

	fmt.Println("System Online. Waiting for MQTT messages...")
	select {}
}