package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Scoped is a GORM scope that automatically filters queries by tenant_id
func Scoped(tenantID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}
