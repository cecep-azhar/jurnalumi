package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Tenant represents a Family Account (Multi-Tenancy Discriminator)
type Tenant struct {
	Base
	Name          string    `gorm:"size:255;not null" json:"name"`
	Plan          string    `gorm:"size:50;default:'free'" json:"plan"` // free, premium
	PlanExpiresAt *time.Time `json:"plan_expires_at"`
	Users         []User    `gorm:"foreignKey:TenantID" json:"users,omitempty"`
}

// User represents family members
type User struct {
	Base
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:50;default:'member'" json:"role"`
}

// Category represents Transaction Categories (Master Data)
type Category struct {
	Base
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Type       string    `gorm:"size:50;not null" json:"type"` // income, expense
	Name       string    `gorm:"size:255;not null" json:"name"`
	Color      string    `gorm:"size:50;default:'gray'" json:"color"`
	BudgetLimit int64   `gorm:"type:bigint;default:0.00" json:"budget_limit"` // Phase 3: Budget Capping
}

// Wallet represents Bank, Cash, or E-Wallet accounts (and Sinking Funds)
type Wallet struct {
	Base
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Type         string    `gorm:"size:50;not null" json:"type"` 
	Balance      int64   `gorm:"type:bigint;default:0.00" json:"balance"`
	TargetAmount int64   `gorm:"type:bigint;default:0.00" json:"target_amount"` // Sinking/Emergency Fund Target
}

// Transaction represents financial ledger items
type Transaction struct {
	Base
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	WalletID        uuid.UUID `gorm:"type:uuid;not null;index" json:"wallet_id"`
	CategoryID      *uuid.UUID `gorm:"type:uuid;index" json:"category_id"` // Link to Category DB
	Type            string    `gorm:"size:50;not null" json:"type"` 
	CategoryName    string    `gorm:"size:100;not null" json:"category_name"` // Legacy / Denormalized
	Amount          int64   `gorm:"type:bigint;not null" json:"amount"`
	Description     string    `gorm:"type:text" json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}

// CommodityAsset represents Gold, Dinar, Perak
type CommodityAsset struct {
	Base
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Type         string    `gorm:"size:50;not null" json:"type"` 
	Name         string    `gorm:"size:255;not null" json:"name"`
	WeightGram   int64   `gorm:"type:bigint;not null" json:"weight_gram"`
	Karatage     int64   `gorm:"type:bigint;default:24.00" json:"karatage"`
	BuyPrice     int64   `gorm:"type:bigint;not null" json:"buy_price"`
	CurrentValue int64   `gorm:"type:bigint;default:0.00" json:"current_value"`
}

// Debt represents Utang and Piutang
type Debt struct {
	Base
	TenantID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Type            string     `gorm:"size:50;not null" json:"type"` 
	Title           string     `gorm:"size:255;not null" json:"title"`
	Counterparty    string     `gorm:"size:255;not null" json:"counterparty"`
	TotalAmount     int64    `gorm:"type:bigint;not null" json:"total_amount"`
	RemainingAmount int64    `gorm:"type:bigint;not null" json:"remaining_amount"`
	InterestRate    int64    `gorm:"type:bigint;default:0.00" json:"interest_rate"`
	DueDate         *time.Time `json:"due_date"`
	Status          string     `gorm:"size:50;default:'active'" json:"status"` 
}

// Voucher represents promotional / activation codes
type Voucher struct {
	Base
	Code      string     `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Duration  int        `gorm:"not null;default:30" json:"duration_days"` 
	IsUsed    bool       `gorm:"default:false" json:"is_used"`
	UsedBy    *uuid.UUID `gorm:"type:uuid" json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
}

// Budget represents monthly budget per category
type Budget struct {
	Base
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Period     string    `gorm:"size:7;not null" json:"period"` // format: YYYY-MM
	Amount     int64     `gorm:"type:bigint;not null" json:"amount"`
}

// PriceSnapshot represents historical price data
type PriceSnapshot struct {
	Base
	CommodityType string    `gorm:"size:50;not null;index" json:"commodity_type"`
	PricePerGram  int64     `gorm:"type:bigint;not null" json:"price_per_gram"`
	Source        string    `gorm:"size:100;not null" json:"source"`
	SnapshotDate  time.Time `gorm:"index" json:"snapshot_date"`
}

// Payment represents subscription payments
type Payment struct {
	Base
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ExternalID    string     `gorm:"size:100;uniqueIndex;not null" json:"external_id"`
	Amount        int64      `gorm:"type:bigint;not null" json:"amount"`
	Status        string     `gorm:"size:50;not null" json:"status"`
	PaymentMethod string     `gorm:"size:50" json:"payment_method"`
	PaidAt        *time.Time `json:"paid_at"`
}

// CategoryWithBudget extends Category with realized budget info for UI
type CategoryWithBudget struct {
	Category
	RealizedAmount int64
	BudgetAmount   int64
	Percentage     int
}
