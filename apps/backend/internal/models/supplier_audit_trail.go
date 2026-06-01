package models

import (
	"time"
)

// SupplierAuditTrail represents the append-only audit trail for all supplier transactions
// Story 10.7: Implement Supplier Transaction Audit Trail
// This model ensures Badan POM compliance with complete 4 W's logging (Who, When, What, Why)
type SupplierAuditTrail struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	TransactionType     string    `json:"transactionType" gorm:"column:transaction_type;not null"`
	EntityType            string    `json:"entityType" gorm:"column:entity_type;not null"`
	EntityID              uint      `json:"entityId" gorm:"column:entity_id;not null"`
	UserID                uint      `json:"userId" gorm:"column:user_id;not null"`
	UserRole              string    `json:"userRole" gorm:"column:user_role;not null"`
	ActionType            string    `json:"actionType" gorm:"column:action_type;not null"`
	ActionDescription      string    `json:"actionDescription" gorm:"column:action_description;not null"`
	Reason                string    `json:"reason,omitempty" gorm:"column:reason"`
	TransactionAmount     *float64  `json:"transactionAmount,omitempty" gorm:"column:transaction_amount"`
	AffectedItemsCount    int       `json:"affectedItemsCount" gorm:"column:affected_items_count;default:0"`
	IPAddress             string    `json:"ipAddress,omitempty" gorm:"column:ip_address"`
	UserAgent             string    `json:"userAgent,omitempty" gorm:"column:user_agent"`
	BranchID              uint      `json:"branchId" gorm:"column:branch_id;not null"`
	CreatedAt             time.Time `json:"createdAt" gorm:"column:created_at;not null"`
}

// TableName specifies the table name for GORM
func (SupplierAuditTrail) TableName() string {
	return "supplier_audit_trail"
}
