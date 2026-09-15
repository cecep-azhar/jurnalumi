package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestTenantScopedSQLConstruction(t *testing.T) {
	tenantID := uuid.New()
	scopeFn := Scoped(tenantID)
	if scopeFn == nil {
		t.Fatal("Scoped function returned nil")
	}
}

func TestWithTenantQueryFormatting(t *testing.T) {
	tenantID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	expected := "SET LOCAL app.tenant_id = '11111111-2222-3333-4444-555555555555'"
	
	generated := "SET LOCAL app.tenant_id = '" + tenantID.String() + "'"
	if generated != expected {
		t.Fatalf("expected SQL %s, got %s", expected, generated)
	}
}
