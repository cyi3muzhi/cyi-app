package biz

import (
	"context"
	"io"
	"testing"

	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

type mockStaffRepo struct {
	listFn func(ctx context.Context, page, pageSize int32, keyword, status, role string) ([]*Staff, int32, error)
}

func (m *mockStaffRepo) List(ctx context.Context, page, pageSize int32, keyword, status, role string) ([]*Staff, int32, error) {
	return m.listFn(ctx, page, pageSize, keyword, status, role)
}

func (m *mockStaffRepo) Create(ctx context.Context, s *Staff, password string) (int64, error) {
	return 0, nil
}

func (m *mockStaffRepo) Update(ctx context.Context, s *Staff) error {
	return nil
}

func (m *mockStaffRepo) ToggleStatus(ctx context.Context, id int64) (int, error) {
	return 0, nil
}

// TestStaffUseCase_List_PageValidation 测试页码校验逻辑：page < 1 修正为 1，page >= 1 原样传递
func TestStaffUseCase_List_PageValidation(t *testing.T) {
	tests := []struct {
		name      string
		page      int32
		wantPage  int32
	}{
		{"页码为零修正为1", 0, 1},
		{"页码为负数修正为1", -5, 1},
		{"页码为1原样传递", 1, 1},
		{"页码大于1原样传递", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPage int32
			uc := NewStaffUseCase(&mockStaffRepo{
				listFn: func(_ context.Context, page, _ int32, _, _, _ string) ([]*Staff, int32, error) {
					gotPage = page
					return nil, 0, nil
				},
			}, kratoslog.NewStdLogger(io.Discard))

			_, _, _ = uc.List(context.Background(), tt.page, 10, "", "", "")
			assert.Equal(t, tt.wantPage, gotPage)
		})
	}
}

// TestStaffUseCase_List_PageSizeValidation 测试每页条数校验逻辑：pageSize < 1 或 > 100 修正为 20，合法值原样传递
func TestStaffUseCase_List_PageSizeValidation(t *testing.T) {
	tests := []struct {
		name         string
		pageSize     int32
		wantPageSize int32
	}{
		{"每页条数为零修正为20", 0, 20},
		{"每页条数为负数修正为20", -1, 20},
		{"每页条数超过100修正为20", 101, 20},
		{"每页条数为1原样传递", 1, 1},
		{"每页条数为50原样传递", 50, 50},
		{"每页条数为100原样传递", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPageSize int32
			uc := NewStaffUseCase(&mockStaffRepo{
				listFn: func(_ context.Context, _ int32, pageSize int32, _, _, _ string) ([]*Staff, int32, error) {
					gotPageSize = pageSize
					return nil, 0, nil
				},
			}, kratoslog.NewStdLogger(io.Discard))

			_, _, _ = uc.List(context.Background(), 1, tt.pageSize, "", "", "")
			assert.Equal(t, tt.wantPageSize, gotPageSize)
		})
	}
}

// TestStaffUseCase_List_RepoError 测试当 repo 返回错误时，错误能正确向上传播
func TestStaffUseCase_List_RepoError(t *testing.T) {
	uc := NewStaffUseCase(&mockStaffRepo{
		listFn: func(_ context.Context, _, _ int32, _, _, _ string) ([]*Staff, int32, error) {
			return nil, 0, io.ErrUnexpectedEOF
		},
	}, kratoslog.NewStdLogger(io.Discard))

	_, _, err := uc.List(context.Background(), 1, 10, "", "", "")
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

// TestStaffUseCase_List_Success 测试正常调用时修正后的分页参数和筛选条件正确传递给 repo
func TestStaffUseCase_List_Success(t *testing.T) {
	var (
		gotPage     int32
		gotPageSize int32
		gotKeyword  string
		gotStatus   string
		gotRole     string
	)
	staffs := []*Staff{{ID: 1, Name: "test"}}
	uc := NewStaffUseCase(&mockStaffRepo{
		listFn: func(_ context.Context, page, pageSize int32, keyword, status, role string) ([]*Staff, int32, error) {
			gotPage, gotPageSize, gotKeyword, gotStatus, gotRole = page, pageSize, keyword, status, role
			return staffs, 1, nil
		},
	}, kratoslog.NewStdLogger(io.Discard))

	result, total, err := uc.List(context.Background(), 0, 200, "张三", "active", "admin")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), gotPage)
	assert.Equal(t, int32(20), gotPageSize)
	assert.Equal(t, "张三", gotKeyword)
	assert.Equal(t, "active", gotStatus)
	assert.Equal(t, "admin", gotRole)
	assert.Equal(t, staffs, result)
	assert.Equal(t, int32(1), total)
}
