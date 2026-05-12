package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBranchTableName(t *testing.T) {
	branch := Branch{}
	assert.Equal(t, "branches", branch.TableName())
}

func TestBranchJSONSerialization(t *testing.T) {
	branch := Branch{
		ID:      1,
		Name:    "Jakarta Branch",
		Address: "Jl. Sudirman No. 1",
		Phone:   "02112345678",
		Email:   "jakarta@simpo.com",
	}
	jsonBytes, err := json.Marshal(branch)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"name":"Jakarta Branch"`)
	assert.Contains(t, string(jsonBytes), `"address":"Jl. Sudirman No. 1"`)
}

func TestBranchHasRelationships(t *testing.T) {
	// Test that Branch struct can hold relationships
	branch := Branch{
		Products:     []Product{},
		Transactions: []Transaction{},
	}
	assert.NotNil(t, branch.Products)
	assert.NotNil(t, branch.Transactions)
}
