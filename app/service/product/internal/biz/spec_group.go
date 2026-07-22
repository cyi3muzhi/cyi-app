package biz

import "context"

// SpecGroup 规格组业务实体
// 对应商品的一组规格名称，如"颜色"、"尺码"
type SpecGroup struct {
	ID        int64
	ProductID int64
	Name      string
	Values    []string // 规格值列表，如 ["红色","蓝色"]
}

// SpecGroupRepo 规格组仓储接口
type SpecGroupRepo interface {
	// CreateSpecGroups 批量创建规格组及其值
	CreateSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error
	// ReplaceSpecGroups 替换规格组（先删后插）
	ReplaceSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error
	// GetSpecGroupsByProductID 获取商品的所有规格组
	GetSpecGroupsByProductID(ctx context.Context, productID int64) ([]*SpecGroup, error)
	// GetSpecGroupsByProductIDs 批量获取多个商品的规格组，返回 map[productID][]*SpecGroup
	GetSpecGroupsByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*SpecGroup, error)
	// DeleteSpecGroupsByProductID 删除商品的所有规格组
	DeleteSpecGroupsByProductID(ctx context.Context, productID int64) error
}

// SpecGroupUseCase 规格组用例
type SpecGroupUseCase struct {
	repo SpecGroupRepo
}

// NewSpecGroupUseCase 创建规格组用例
func NewSpecGroupUseCase(repo SpecGroupRepo) *SpecGroupUseCase {
	return &SpecGroupUseCase{repo: repo}
}

// CreateSpecGroups 批量创建规格组
func (uc *SpecGroupUseCase) CreateSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error {
	return uc.repo.CreateSpecGroups(ctx, productID, groups)
}

// ReplaceSpecGroups 替换规格组
func (uc *SpecGroupUseCase) ReplaceSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error {
	return uc.repo.ReplaceSpecGroups(ctx, productID, groups)
}

// GetSpecGroupsByProductID 获取商品的规格组
func (uc *SpecGroupUseCase) GetSpecGroupsByProductID(ctx context.Context, productID int64) ([]*SpecGroup, error) {
	return uc.repo.GetSpecGroupsByProductID(ctx, productID)
}

// DeleteSpecGroupsByProductID 删除商品的规格组
func (uc *SpecGroupUseCase) DeleteSpecGroupsByProductID(ctx context.Context, productID int64) error {
	return uc.repo.DeleteSpecGroupsByProductID(ctx, productID)
}
