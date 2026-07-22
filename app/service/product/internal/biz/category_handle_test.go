package biz

import (
	"context"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockCategoryRepo 是 CategoryRepo 的模拟实现，仅 ListCategoryTree 使用 listFn，其余方法返回零值。
type mockCategoryRepo struct {
	listFn func(ctx context.Context, parentID int64) ([]*Category, error)
}

func (m *mockCategoryRepo) ListCategoryTree(ctx context.Context, parentID int64) ([]*Category, error) {
	return m.listFn(ctx, parentID)
}
func (m *mockCategoryRepo) GetCategory(_ context.Context, _ int64) (*Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) CreateCategory(_ context.Context, _ *Category) error    { return nil }
func (m *mockCategoryRepo) UpdateCategory(_ context.Context, _ *Category) error    { return nil }
func (m *mockCategoryRepo) DeleteCategory(_ context.Context, _ int64) error        { return nil }
func (m *mockCategoryRepo) ListChildCategories(_ context.Context, _ int64) ([]*Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) UpdateCategorySort(_ context.Context, _ int64, _ int32) error {
	return nil
}
func (m *mockCategoryRepo) UpdateCategoryStatus(_ context.Context, _ int64, _ bool) error {
	return nil
}
func (m *mockCategoryRepo) ToggleCategoryShowcase(_ context.Context, _ int64, _ bool) error {
	return nil
}
func (m *mockCategoryRepo) BatchOperation(_ context.Context, _ string, _ []int64) (int32, error) {
	return 0, nil
}

// newTestUseCase 根据传入的 listFn 创建一个使用模拟仓库的 CategoryUseCase。
func newTestUseCase(listFn func(ctx context.Context, parentID int64) ([]*Category, error)) *CategoryUseCase {
	return NewCategoryUseCase(&mockCategoryRepo{listFn: listFn}, log.NewStdLogger(io.Discard))
}

// TestBuildTree_FullTree 验证从扁平列表构建完整树结构，ParentID==0 为根节点，Children 正确嵌套。
func TestBuildTree_FullTree(t *testing.T) {
	uc := newTestUseCase(nil)
	all := []*Category{
		{ID: 1, ParentID: 0, Name: "电子"},
		{ID: 2, ParentID: 0, Name: "服装"},
		{ID: 3, ParentID: 1, Name: "手机"},
		{ID: 4, ParentID: 1, Name: "电脑"},
		{ID: 5, ParentID: 2, Name: "男装"},
	}
	tree := uc.buildTree(all, 0)

	assert.Len(t, tree, 2)
	assert.Equal(t, int64(1), tree[0].ID)
	assert.Len(t, tree[0].Children, 2)
	assert.Equal(t, int64(3), tree[0].Children[0].ID)
	assert.Equal(t, int64(4), tree[0].Children[1].ID)
	assert.Len(t, tree[1].Children, 1)
	assert.Equal(t, int64(5), tree[1].Children[0].ID)
}

// TestBuildTree_WithParentID 验证当 parentID > 0 时，仅返回指定父节点下的子树。
func TestBuildTree_WithParentID(t *testing.T) {
	uc := newTestUseCase(nil)
	all := []*Category{
		{ID: 1, ParentID: 0, Name: "电子"},
		{ID: 2, ParentID: 1, Name: "手机"},
		{ID: 3, ParentID: 1, Name: "电脑"},
		{ID: 4, ParentID: 2, Name: "智能手机"},
		{ID: 5, ParentID: 0, Name: "服装"},
	}
	tree := uc.buildTree(all, 1)

	assert.Len(t, tree, 2)
	assert.Equal(t, int64(2), tree[0].ID)
	assert.Equal(t, int64(3), tree[1].ID)
	assert.Len(t, tree[0].Children, 1)
	assert.Equal(t, int64(4), tree[0].Children[0].ID)
}

// TestBuildTree_EmptyList 验证空输入返回空树。
func TestBuildTree_EmptyList(t *testing.T) {
	uc := newTestUseCase(nil)
	tree := uc.buildTree(nil, 0)
	assert.Empty(t, tree)
}

// TestBuildTree_SingleRoot 验证单个根节点无子节点的情况。
func TestBuildTree_SingleRoot(t *testing.T) {
	uc := newTestUseCase(nil)
	all := []*Category{
		{ID: 1, ParentID: 0, Name: "唯一"},
	}
	tree := uc.buildTree(all, 0)

	assert.Len(t, tree, 1)
	assert.Equal(t, int64(1), tree[0].ID)
	assert.Empty(t, tree[0].Children)
}

// TestBuildTree_DeepNesting 验证三层深度嵌套（祖父 → 父 → 子）。
func TestBuildTree_DeepNesting(t *testing.T) {
	uc := newTestUseCase(nil)
	all := []*Category{
		{ID: 1, ParentID: 0, Name: "祖父"},
		{ID: 2, ParentID: 1, Name: "父"},
		{ID: 3, ParentID: 2, Name: "子"},
	}
	tree := uc.buildTree(all, 0)

	assert.Len(t, tree, 1)
	assert.Equal(t, int64(1), tree[0].ID)
	assert.Len(t, tree[0].Children, 1)
	assert.Equal(t, int64(2), tree[0].Children[0].ID)
	assert.Len(t, tree[0].Children[0].Children, 1)
	assert.Equal(t, int64(3), tree[0].Children[0].Children[0].ID)
}

// TestSortBySort 验证 sortBySort 按 Sort 字段升序排列。
func TestSortBySort(t *testing.T) {
	uc := newTestUseCase(nil)
	list := []*Category{
		{ID: 1, Sort: 30},
		{ID: 2, Sort: 10},
		{ID: 3, Sort: 20},
	}
	uc.sortBySort(list)

	assert.Equal(t, int32(10), list[0].Sort)
	assert.Equal(t, int32(20), list[1].Sort)
	assert.Equal(t, int32(30), list[2].Sort)
}

// TestSortBySort_EmptyAndSingle 验证空列表和单元素列表排序不会崩溃。
func TestSortBySort_EmptyAndSingle(t *testing.T) {
	uc := newTestUseCase(nil)

	// 空列表
	var empty []*Category
	uc.sortBySort(empty)
	assert.Empty(t, empty)

	// 单元素
	single := []*Category{{ID: 1, Sort: 5}}
	uc.sortBySort(single)
	assert.Len(t, single, 1)
	assert.Equal(t, int32(5), single[0].Sort)
}

// TestSortBySort_AlreadySorted 验证已排序列表排序后顺序不变。
func TestSortBySort_AlreadySorted(t *testing.T) {
	uc := newTestUseCase(nil)
	list := []*Category{
		{ID: 1, Sort: 1},
		{ID: 2, Sort: 2},
		{ID: 3, Sort: 3},
	}
	uc.sortBySort(list)

	assert.Equal(t, int32(1), list[0].Sort)
	assert.Equal(t, int32(2), list[1].Sort)
	assert.Equal(t, int32(3), list[2].Sort)
}

// TestListCategoryTree_RepoError 验证当仓库返回错误时，ListCategoryTree 能正确传播错误。
func TestListCategoryTree_RepoError(t *testing.T) {
	wantErr := assert.AnError
	uc := newTestUseCase(func(_ context.Context, _ int64) ([]*Category, error) {
		return nil, wantErr
	})

	got, err := uc.ListCategoryTree(context.Background(), 0)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, wantErr)
}

// TestDeleteCategory_PropagatesProductConflict 验证当分类下存在商品时，DeleteCategory 应向上传递
// "分类下存在商品"错误，避免数据层在缺少前置校验时把分类连同其商品一并删除，导致 product.category
// 字段悬空、用户无法通过分类导航找到这些商品。
func TestDeleteCategory_PropagatesProductConflict(t *testing.T) {
	wantErr := assert.AnError
	uc := &CategoryUseCase{
		repo: &deleteCategoryRepo{
			deleteFn: func(_ context.Context, _ int64) error { return wantErr },
		},
		log: log.NewHelper(log.NewStdLogger(io.Discard)),
	}

	err := uc.DeleteCategory(context.Background(), 42)
	assert.ErrorIs(t, err, wantErr)
}

// deleteCategoryRepo 仅用于验证 DeleteCategory 错误传播的最小化模拟实现
type deleteCategoryRepo struct {
	deleteFn func(ctx context.Context, id int64) error
}

func (d *deleteCategoryRepo) ListCategoryTree(_ context.Context, _ int64) ([]*Category, error) {
	return nil, nil
}
func (d *deleteCategoryRepo) GetCategory(_ context.Context, _ int64) (*Category, error) { return nil, nil }
func (d *deleteCategoryRepo) CreateCategory(_ context.Context, _ *Category) error      { return nil }
func (d *deleteCategoryRepo) UpdateCategory(_ context.Context, _ *Category) error      { return nil }
func (d *deleteCategoryRepo) DeleteCategory(ctx context.Context, id int64) error       { return d.deleteFn(ctx, id) }
func (d *deleteCategoryRepo) ListChildCategories(_ context.Context, _ int64) ([]*Category, error) {
	return nil, nil
}
func (d *deleteCategoryRepo) UpdateCategorySort(_ context.Context, _ int64, _ int32) error  { return nil }
func (d *deleteCategoryRepo) UpdateCategoryStatus(_ context.Context, _ int64, _ bool) error { return nil }
func (d *deleteCategoryRepo) ToggleCategoryShowcase(_ context.Context, _ int64, _ bool) error {
	return nil
}
func (d *deleteCategoryRepo) BatchOperation(_ context.Context, _ string, _ []int64) (int32, error) {
	return 0, nil
}
