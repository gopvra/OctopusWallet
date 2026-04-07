package models

import "time"

const (
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

type AdminUser struct {
	ID        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	Role      string    `db:"role" json:"role"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ChainState struct {
	Chain            string    `db:"chain" json:"chain"`
	LastScannedBlock uint64    `db:"last_scanned_block" json:"last_scanned_block"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
