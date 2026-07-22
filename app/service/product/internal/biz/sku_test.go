package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkuUseCase_CreateSkus(t *testing.T) {
	var called bool
	repo := &mockSkuRepo{
		createSkus: func(ctx context.Context, skus []*Sku) error {
			called = true
			assert.Len(t, skus, 2)
			return nil
		},
	}
	uc := NewSkuUseCase(repo)

	err := uc.CreateSkus(context.Background(), []*Sku{{Price: 100}, {Price: 200}})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSkuUseCase_ReplaceSkus(t *testing.T) {
	var called bool
	repo := &mockSkuRepo{
		replaceSkus: func(ctx context.Context, productID int64, skus []*Sku) error {
			called = true
			assert.Equal(t, int64(1), productID)
			assert.Len(t, skus, 2)
			return nil
		},
	}
	uc := NewSkuUseCase(repo)

	err := uc.ReplaceSkus(context.Background(), 1, []*Sku{{Price: 100}, {Price: 200}})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSkuUseCase_GetSkusByProductID(t *testing.T) {
	var called bool
	repo := &mockSkuRepo{
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			called = true
			assert.Equal(t, int64(1), productID)
			return []*Sku{{ProductID: 1, Price: 100}, {ProductID: 1, Price: 200}}, nil
		},
	}
	uc := NewSkuUseCase(repo)

	skus, err := uc.GetSkusByProductID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, skus, 2)
	assert.True(t, called)
}

func TestSkuUseCase_DeleteSkusByProductID(t *testing.T) {
	var called bool
	repo := &mockSkuRepo{
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			called = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	uc := NewSkuUseCase(repo)

	err := uc.DeleteSkusByProductID(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, called)
}