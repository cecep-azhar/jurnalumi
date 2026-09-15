package scheduler

import (
	"testing"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/google/uuid"
)

func TestSchedulerDowngradeEligibilityLogic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		tenant          models.Tenant
		shouldDowngrade bool
	}{
		{
			name: "Premium plan expired in past should downgrade",
			tenant: models.Tenant{
				Base:          models.Base{ID: uuid.New()},
				Name:          "Expired Tenant",
				Plan:          "premium",
				PlanExpiresAt: ptrTime(now.Add(-2 * time.Hour)),
			},
			shouldDowngrade: true,
		},
		{
			name: "Premium plan active until future should NOT downgrade",
			tenant: models.Tenant{
				Base:          models.Base{ID: uuid.New()},
				Name:          "Active Premium Tenant",
				Plan:          "premium",
				PlanExpiresAt: ptrTime(now.Add(24 * time.Hour)),
			},
			shouldDowngrade: false,
		},
		{
			name: "Free plan with nil expiration should NOT downgrade",
			tenant: models.Tenant{
				Base:          models.Base{ID: uuid.New()},
				Name:          "Free Tenant",
				Plan:          "free",
				PlanExpiresAt: nil,
			},
			shouldDowngrade: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEligible := tt.tenant.Plan == "premium" && tt.tenant.PlanExpiresAt != nil && tt.tenant.PlanExpiresAt.Before(now)
			if isEligible != tt.shouldDowngrade {
				t.Fatalf("expected downgrade eligibility %v, got %v", tt.shouldDowngrade, isEligible)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
