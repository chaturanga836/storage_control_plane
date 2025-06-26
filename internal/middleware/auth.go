// internal/middleware/auth.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/auth"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/models"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils/contextkeys"
)

// ✅ Parses JWT and attaches claims to request context
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		log.Printf("🪪 Incoming Auth Header: %s", authHeader)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Printf("✅ JWT validated for user: %s with role: %s", claims.Username, claims.Role)
		ctx := context.WithValue(r.Context(), contextkeys.UserClaimsKey, *claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ✅ Ensures only super admins can access protected routes
func RequireSuperAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextkeys.UserClaimsKey).(models.UserClaims)
		if !ok || !claims.IsSuperAdmin {
			log.Printf("⛔ Access denied – not super admin. Claims: %+v", claims)
			http.Error(w, "Forbidden – Super admin only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}


