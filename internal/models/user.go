// internal/models/user.go
package models

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // hashed
	Role     string `json:"role"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}


// ✅ Used by middleware/auth.go to inject claims into context
type UserClaims struct {
	Username     string `json:"username"`
	Role         string `json:"role"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}