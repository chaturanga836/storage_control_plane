package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Node struct {
	ID           string   `json:"id"`
	Hostname     string   `json:"hostname"`
	IPAddress    string   `json:"ip_address"`
	Status       string   `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
	StorageUsed  int64    `json:"storage_used"`
	TotalStorage int64    `json:"total_storage"`
	Tags         []string `json:"tags"`
}

var nodeRegistry = make(map[string]Node)

func RegisterNode(w http.ResponseWriter, r *http.Request) {
	var node Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	node.LastSeen = time.Now()
	nodeRegistry[node.ID] = node

	fmt.Printf("📥 Registered node: %s (%s)\n", node.ID, node.IPAddress)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Node registered"})
}

type HeartbeatPayload struct {
	ID          string `json:"id"`
	StorageUsed int64  `json:"storage_used"`
}

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	var hb HeartbeatPayload
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	node, exists := nodeRegistry[hb.ID]
	if !exists {
		http.Error(w, "Unknown node", http.StatusNotFound)
		return
	}

	node.StorageUsed = hb.StorageUsed
	node.LastSeen = time.Now()
	nodeRegistry[hb.ID] = node

	fmt.Printf("♥ Heartbeat from %s | Used: %d MB\n", hb.ID, hb.StorageUsed/1024/1024)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Heartbeat received"})
}
