// models/vendor.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Vendor struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"uniqueIndex" json:"user_id"` // linked to users table
	ShopName   string         `gorm:"size:255" json:"shop_name"`
	Status     string         `gorm:"size:50;default:'pending'" json:"status"` // pending/active/suspended/rejected
	ApprovedBy *uint          `json:"approved_by"`                             // user id of super-admin who approved
	ApprovedAt *time.Time     `json:"approved_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// models/audit_log.go

type AuditLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ActorID   *uint          `json:"actor_id"` // who did the action (nullable for system)
	Action    string         `gorm:"size:200" json:"action"`
	Resource  string         `gorm:"size:200" json:"resource"` // e.g., "vendor:123"
	OldValue  string         `gorm:"type:text" json:"old_value"`
	NewValue  string         `gorm:"type:text" json:"new_value"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
