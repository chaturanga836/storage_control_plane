// internal/logger/audit.go 
package logger

import (
	"encoding/json"
	"os"
	"time"
)

type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	Route     string `json:"route"`
	Status    string `json:"status"`
	IP        string `json:"ip"`
}

func LogAudit(username, route, status, ip string) {
	entry := AuditEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Username:  username,
		Route:     route,
		Status:    status,
		IP:        ip,
	}

	file, err := os.OpenFile("logs/audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Optional: log to stdout instead
	}
	defer file.Close()

	json.NewEncoder(file).Encode(entry)
}
