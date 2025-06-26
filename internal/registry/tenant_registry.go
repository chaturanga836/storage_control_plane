// internal/registry/tenant_registry.go
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"context"
	"time"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/models"
)

var (
	tenantList []models.Tenant
	mu         sync.RWMutex
	dataFile   = "data/tenants.json"
)

// LoadTenantRegistry loads tenants from file into memory
func LoadTenantRegistry() error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			tenantList = []models.Tenant{}
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&tenantList)
}

// SaveTenantRegistry writes the in-memory tenant list to file
func SaveTenantRegistry() error {
	mu.RLock()
	defer mu.RUnlock()

	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tenantList)
}

// RegisterTenant adds a new tenant if not already present
func RegisterTenant(t models.Tenant) error {
	mu.Lock()
	defer mu.Unlock()

	for _, tenant := range tenantList {
		if tenant.TenantID == t.TenantID {
			return fmt.Errorf("tenant TenantID already exists")
		}
	}

	tenantList = append(tenantList, t)
	return SaveTenantRegistry()
}

// AssignNodeToTenant updates the node for an existing tenant
func AssignNodeToTenant(tenantID, nodeID string) error {
	mu.Lock()
	defer mu.Unlock()

	for i, t := range tenantList {
		if t.TenantID == tenantID {
			tenantList[i].NodeID = nodeID
			return SaveTenantRegistry()
		}
	}
	return fmt.Errorf("tenant not found")
}

// GetTenantByID looks up a tenant by ID
func GetTenantByID(id string) (models.Tenant, bool) {
	mu.RLock()
	defer mu.RUnlock()

	for _, t := range tenantList {
		if t.TenantID == id {
			return t, true
		}
	}
	return models.Tenant{}, false
}

// ReloadTenants reloads the tenant list from disk
func ReloadTenants() error {
	return LoadTenantRegistry()
}

// GetAllTenants returns the entire tenant list
func GetAllTenants() []*models.Tenant {
	mu.RLock()
	defer mu.RUnlock()

	list := make([]*models.Tenant, 0, len(tenantList))
	for i := range tenantList {
		list = append(list, &tenantList[i])
	}
	return list
}

func WatchTenantFileChanges(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := ReloadTenants(); err != nil {
				fmt.Printf("⚠️ Tenant reload failed: %v\n", err)
			}
		}
	}
}
