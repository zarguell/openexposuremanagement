package auth

import (
	"net/http"
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleAnalyst = "analyst"
	RoleViewer = "viewer"
)

// RoleRequirement specifies required roles for endpoint access
type RoleRequirement struct {
	Roles []string
	Any   bool // If true, user needs any of the roles; if false, needs all
}

// RequireRoles creates middleware that checks for required roles
func RequireRoles(requirement RoleRequirement) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Extract user context from request
			// For MVP, skip role checks
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin creates middleware that requires admin role
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRoles(RoleRequirement{
		Roles: []string{RoleAdmin},
		Any:   true,
	})
}

// RequireAnalyst creates middleware that requires admin or analyst role
func RequireAnalyst() func(http.Handler) http.Handler {
	return RequireRoles(RoleRequirement{
		Roles: []string{RoleAdmin, RoleAnalyst},
		Any:   true,
	})
}

// RequireViewer creates middleware that requires any role (viewer+)
func RequireViewer() func(http.Handler) http.Handler {
	return RequireRoles(RoleRequirement{
		Roles: []string{RoleAdmin, RoleAnalyst, RoleViewer},
		Any:   true,
	})
}
