// internal/registry/user_registry.go
package registry

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/models"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
)

// 🔄 Efficient in-memory user map: username -> *User
var (
	userIndex = make(map[string]*models.User)
	userMu    sync.RWMutex
)

// LoadUserRegistry loads and indexes users.json into memory
func LoadUserRegistry() {
	loaded, err := utils.LoadUsers()
	if err != nil {
		log.Fatalf("❌ Failed to load users: %v", err)
	}

	temp := make(map[string]*models.User)
	for i := range loaded {
		u := loaded[i]
		temp[u.Username] = &u
	}

	userMu.Lock()
	userIndex = temp
	userMu.Unlock()

	log.Printf("👥 Loaded %d users into registry", len(userIndex))
}

// ReloadUsers safely reloads users.json
func ReloadUsers() {
	LoadUserRegistry()
}

// GetUserByUsername provides fast lookup
func GetUserByUsername(username string) (*models.User, bool) {
	userMu.RLock()
	defer userMu.RUnlock()
	u, ok := userIndex[username]
	return u, ok
}

// WatchUserFileChanges periodically reloads users.json
func WatchUserFileChanges(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ReloadUsers()
		}
	}
}
