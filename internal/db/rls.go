package db

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WithTenant sets the PostgreSQL app.tenant_id variable for Row-Level Security
// within a transaction. MUST be used with a transaction block.
func WithTenant(tx *gorm.DB, tenantID uuid.UUID) *gorm.DB {
	tx.Exec(fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID.String()))
	return tx
}

// EnableRLS enables Row Level Security on a specific table
func EnableRLS(db *gorm.DB, tableName string) error {
	policyName := fmt.Sprintf("tenant_isolation_policy_%s", tableName)
	
	// Force owner/superadmin to bypass RLS to prevent locked out
	db.Exec(fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY;", tableName))
	
	if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY;", tableName)).Error; err != nil {
		return err
	}
	
	// Drop existing policy if exists to recreate
	db.Exec(fmt.Sprintf("DROP POLICY IF EXISTS %s ON %s;", policyName, tableName))
	
	// Create policy
	policySQL := fmt.Sprintf(`
		CREATE POLICY %s ON %s
		USING (
			current_setting('app.tenant_id', true) IS NULL OR current_setting('app.tenant_id', true) = ''
			OR tenant_id = current_setting('app.tenant_id', true)::uuid
		);
	`, policyName, tableName)
	
	return db.Exec(policySQL).Error
}
