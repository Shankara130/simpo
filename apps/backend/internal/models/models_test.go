package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProductTableName(t *testing.T) {
	product := Product{}
	assert.Equal(t, "products", product.TableName())
}

func TestProductJSONSerialization(t *testing.T) {
	price := "75000.00"
	branchID := uint(1)
	expiryDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	product := Product{
		ID:           1,
		SKU:          "TEST123",
		Name:         "Paracetamol 500mg",
		StockQty:     50,
		Price:        price,
		BranchID:     branchID,
		ExpiryDate:   &expiryDate,
		Category:     "OBAT",
	}
	jsonBytes, err := json.Marshal(product)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"sku":"TEST123"`)
	assert.Contains(t, string(jsonBytes), `"stockQty":50`)
	assert.Contains(t, string(jsonBytes), `"price":"75000.00"`)
	assert.Contains(t, string(jsonBytes), `"branchId":1`)
}

func TestTransactionTableName(t *testing.T) {
	transaction := Transaction{}
	assert.Equal(t, "transactions", transaction.TableName())
}

func TestTransactionJSONSerialization(t *testing.T) {
	total := "150000.00"
	cashierID := uint(1)
	branchID := uint(1)
	transaction := Transaction{
		ID:                1,
		TransactionNumber: "TRX-20260512-0001-0001",
		CashierID:         cashierID,
		BranchID:          branchID,
		Total:             total,
		Subtotal:          total,
		PaymentMethod:     PaymentMethodCash,
		Status:            StatusCompleted,
	}
	jsonBytes, err := json.Marshal(transaction)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"transactionNumber":"TRX-20260512-0001-0001"`)
	assert.Contains(t, string(jsonBytes), `"paymentMethod":"CASH"`)
	assert.Contains(t, string(jsonBytes), `"status":"COMPLETED"`)
}

func TestTransactionItemTableName(t *testing.T) {
	item := TransactionItem{}
	assert.Equal(t, "transaction_items", item.TableName())
}

func TestTransactionItemJSONSerialization(t *testing.T) {
	transactionID := uint(1)
	productID := uint(5)
	item := TransactionItem{
		ID:            1,
		TransactionID: transactionID,
		ProductID:     productID,
		Quantity:      2,
		UnitPrice:     "75000.00",
		Subtotal:      "150000.00",
		ProductName:   "Paracetamol 500mg",
		ProductSKU:    "TEST123",
	}
	jsonBytes, err := json.Marshal(item)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"quantity":2`)
	assert.Contains(t, string(jsonBytes), `"unitPrice":"75000.00"`)
	assert.Contains(t, string(jsonBytes), `"productName":"Paracetamol 500mg"`)
}

func TestPaymentMethodConstants(t *testing.T) {
	assert.Equal(t, "CASH", PaymentMethodCash)
	assert.Equal(t, "TRANSFER", PaymentMethodTransfer)
	assert.Equal(t, "E-WALLET", PaymentMethodEWallet)
	assert.Equal(t, "CARD", PaymentMethodCard)
	assert.Equal(t, "QRIS", PaymentMethodQRIS)
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, "PENDING", StatusPending)
	assert.Equal(t, "COMPLETED", StatusCompleted)
	assert.Equal(t, "CANCELLED", StatusCancelled)
	assert.Equal(t, "REFUNDED", StatusRefunded)
}
