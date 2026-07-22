package biz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductStatus_IsOnSale(t *testing.T) {
	assert.True(t, ProductStatusOnSale.IsOnSale())
	assert.False(t, ProductStatusOffSale.IsOnSale())
}

func TestProductStatus_IsOffSale(t *testing.T) {
	assert.False(t, ProductStatusOnSale.IsOffSale())
	assert.True(t, ProductStatusOffSale.IsOffSale())
}

func TestProductStatus_IsValid(t *testing.T) {
	assert.True(t, ProductStatusOnSale.IsValid())
	assert.True(t, ProductStatusOffSale.IsValid())
	assert.False(t, ProductStatus(99).IsValid())
}

func TestProductStatus_String(t *testing.T) {
	assert.Equal(t, "上架", ProductStatusOnSale.String())
	assert.Equal(t, "下架", ProductStatusOffSale.String())
	assert.Equal(t, "未知状态", ProductStatus(99).String())
}

func TestProductStatus_Value(t *testing.T) {
	assert.Equal(t, int32(1), ProductStatusOnSale.Value())
	assert.Equal(t, int32(0), ProductStatusOffSale.Value())
}