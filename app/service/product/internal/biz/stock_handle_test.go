package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStockUseCase_CreateStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		createStock: func(ctx context.Context, in *Stock) error {
			called = true
			assert.Equal(t, int64(1), in.ProductID)
			assert.Equal(t, int64(100), in.SkuID)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.CreateStock(context.Background(), &Stock{ProductID: 1, SkuID: 100})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_UpdateStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		updateStock: func(ctx context.Context, in *Stock) error {
			called = true
			assert.Equal(t, int64(1), in.ID)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.UpdateStock(context.Background(), &Stock{ID: 1})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_ReserveStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		reserveStock: func(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
			called = true
			assert.Equal(t, int64(100), skuID)
			assert.Equal(t, int32(5), count)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.ReserveStock(context.Background(), StockOpKey{}, 100, 5)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_DeductStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		deductStock: func(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
			called = true
			assert.Equal(t, int64(100), skuID)
			assert.Equal(t, int32(5), count)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.DeductStock(context.Background(), StockOpKey{}, 100, 5)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_ReleaseStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		releaseStock: func(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
			called = true
			assert.Equal(t, int64(100), skuID)
			assert.Equal(t, int32(5), count)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.ReleaseStock(context.Background(), StockOpKey{}, 100, 5)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_AddStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		addStock: func(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
			called = true
			assert.Equal(t, int64(100), skuID)
			assert.Equal(t, int32(10), count)
			return nil
		},
	}
	uc := NewStockUseCase(repo)

	err := uc.AddStock(context.Background(), StockOpKey{}, 100, 10)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStockUseCase_GetStock(t *testing.T) {
	var called bool
	repo := &mockStockRepo{
		getStock: func(ctx context.Context, skuID int64) (*Stock, error) {
			called = true
			assert.Equal(t, int64(100), skuID)
			return &Stock{ID: 1, ProductID: 1, SkuID: 100, Total: 100, Available: 80, Reserved: 20}, nil
		},
	}
	uc := NewStockUseCase(repo)

	stock, err := uc.GetStock(context.Background(), 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stock.ID)
	assert.Equal(t, int64(100), stock.SkuID)
	assert.Equal(t, int32(100), stock.Total)
	assert.Equal(t, int32(80), stock.Available)
	assert.Equal(t, int32(20), stock.Reserved)
	assert.True(t, called)
}
