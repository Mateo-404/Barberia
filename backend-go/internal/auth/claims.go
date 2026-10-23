package auth

// Roles de administrador (SPEC §8).
const (
	RolOwner = "owner"
	RolAdmin = "admin"
	RolStaff = "staff"
)

// EsAdminRol indica si el rol puede usar /api/admin/*.
func EsAdminRol(rol string) bool {
	return rol == RolOwner || rol == RolAdmin
}
