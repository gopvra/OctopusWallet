package models

import "time"

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
)

type AdminUser struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"not null;default:'admin'" json:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type ChainState struct {
	Chain            string    `gorm:"primaryKey" json:"chain"`
	LastScannedBlock uint64    `gorm:"not null;default:0" json:"last_scanned_block"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
