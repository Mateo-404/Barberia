package auth

// Admin roles (SPEC §8).
const (
	RoleOwner = "owner"
	RoleAdmin = "admin"
	RoleStaff = "staff"
)

// IsAdminRole reports whether the role may use /api/admin/*.
func IsAdminRole(role string) bool {
	return role == RoleOwner || role == RoleAdmin
}
