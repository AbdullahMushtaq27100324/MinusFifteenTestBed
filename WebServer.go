package main

import(
    "fmt"
    "net/http"
    "encoding/json"
)

// WebServer initializes the admin control panel and HTTP routes
func WebServer() {
    http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
        newID := globalNodeID.Add(1)
        fmt.Printf("--> Admin Command: Spawning Node %d\n", newID)
        startNode(&Node{NodeNumber: int(newID)}, 1) 
        fmt.Fprintf(w, "Success: Node %d is now running!\n", newID)
    })

    http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS
        
        dataMu.Lock()
        response := map[string]interface{}{
        "total_bytes": TotalStorageBytes.Load(),
        "nodes":       latestData,
        }
        json.NewEncoder(w).Encode(response)
        dataMu.Unlock()
    })

    http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Query().Get("id")
        
        var nodeID int
        fmt.Sscanf(idStr, "%d", &nodeID) 

        // Cancel the node's context and remove from active tracking
        activeNodesMu.Lock()
        cancelFunc, exists := activeNodes[nodeID]
        if exists {
            cancelFunc() 
            delete(activeNodes, nodeID) 
            fmt.Printf("--> Admin Command: Killed Node %d\n", nodeID)
        }
        activeNodesMu.Unlock()

        // Clear node data from the dashboard map
        dataMu.Lock()
        delete(latestData, nodeID)
        dataMu.Unlock()
    })

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })

    fmt.Println("Dashboard available at http://localhost:8080")
    http.ListenAndServe(":8080", nil) 
}