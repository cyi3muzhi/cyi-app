package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecGroupUseCase_CreateSpecGroups(t *testing.T) {
	var called bool
	repo := &mockSpecGroupRepo{
		createSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			called = true
			assert.Equal(t, int64(1), productID)
			assert.Len(t, groups, 2)
			return nil
		},
	}
	uc := NewSpecGroupUseCase(repo)

	err := uc.CreateSpecGroups(context.Background(), 1, []*SpecGroup{{Name: "颜色"}, {Name: "尺码"}})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSpecGroupUseCase_ReplaceSpecGroups(t *testing.T) {
	var called bool
	repo := &mockSpecGroupRepo{
		replaceSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			called = true
			assert.Equal(t, int64(1), productID)
			assert.Len(t, groups, 2)
			return nil
		},
	}
	uc := NewSpecGroupUseCase(repo)

	err := uc.ReplaceSpecGroups(context.Background(), 1, []*SpecGroup{{Name: "颜色"}, {Name: "尺码"}})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSpecGroupUseCase_GetSpecGroupsByProductID(t *testing.T) {
	var called bool
	repo := &mockSpecGroupRepo{
		getSpecGroupsByProductID: func(ctx context.Context, productID int64) ([]*SpecGroup, error) {
			called = true
			assert.Equal(t, int64(1), productID)
			return []*SpecGroup{{ProductID: 1, Name: "颜色"}, {ProductID: 1, Name: "尺码"}}, nil
		},
	}
	uc := NewSpecGroupUseCase(repo)

	groups, err := uc.GetSpecGroupsByProductID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.True(t, called)
}

func TestSpecGroupUseCase_DeleteSpecGroupsByProductID(t *testing.T) {
	var called bool
	repo := &mockSpecGroupRepo{
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			called = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	uc := NewSpecGroupUseCase(repo)

	err := uc.DeleteSpecGroupsByProductID(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, called)
}