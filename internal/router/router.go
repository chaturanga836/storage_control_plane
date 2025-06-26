// internal/router/router.go 
package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/handlers"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/middleware"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// ✅ Public routes
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/auth/login", handlers.Login).Methods("POST")
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods("POST")

	// ✅ Protected generic routes
	r.Handle("/auth/me", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.Me))).Methods("GET")
	r.Handle("/protected", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.ProtectedEndpoint))).Methods("GET")

	// ✅ Admin-only protected route (example)
	r.Handle("/admin-only",
		middleware.JWTAuthMiddleware(
			middleware.RequireSuperAdmin(http.HandlerFunc(handlers.AdminOnlyEndpoint)),
		),
	).Methods("GET")

	// ✅ Tenant management routes (protected by super admin)
	r.Handle("/admin/tenants", middleware.JWTAuthMiddleware(middleware.RequireSuperAdmin(http.HandlerFunc(handlers.ListTenants)))).Methods("GET")
	r.Handle("/admin/tenants/register", middleware.JWTAuthMiddleware(middleware.RequireSuperAdmin(http.HandlerFunc(handlers.RegisterTenant)))).Methods("POST")
	r.Handle("/admin/tenants/assign-node", middleware.JWTAuthMiddleware(middleware.RequireSuperAdmin(http.HandlerFunc(handlers.AssignNode)))).Methods("POST")

	// ✅ Monitoring (protected)
	r.Handle("/monitor", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.Monitor))).Methods("GET")

	// ✅ DuckDB inspection routes
	r.Handle("/duckdb/tables", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.ListDuckDBTables))).Methods("GET")
	r.Handle("/duckdb/table/{name}/count", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.GetTableRowCount))).Methods("GET")

	// ✅ Audit log route (super admin only)
	r.Handle("/admin/logs",
		middleware.JWTAuthMiddleware(
			middleware.RequireSuperAdmin(http.HandlerFunc(handlers.GetAuditLogs)),
		),
	).Methods("GET")

	// ✅ Node registration and heartbeat (public for now)
	r.HandleFunc("/register-node", handlers.RegisterNode).Methods("POST")
	r.HandleFunc("/heartbeat", handlers.Heartbeat).Methods("POST")

	// ✅ Fallback for CORS preflight
	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

