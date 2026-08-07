package data

import (
	"context"
	"encoding/json"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
)

type specGroupRepo struct {
	data *Data
	log  *log.Helper
}

// NewSpecGroupRepo 创建规格组仓储
func NewSpecGroupRepo(data *Data, logger log.Logger) biz.SpecGroupRepo {
	return &specGroupRepo{data: data, log: log.NewHelper(logger)}
}

// CreateSpecGroups 批量创建规格组及其值
func (r *specGroupRepo) CreateSpecGroups(ctx context.Context, productID int64, groups []*biz.SpecGroup) error {
	for _, group := range groups {
		// 创建规格组
		specGroup := &mod.SpecGroup{
			ProductID: productID,
			Name:      group.Name,
		}
		if err := curd.SpecGroup.WithContext(ctx).Create(specGroup); err != nil {
			return err
		}
		group.ID = specGroup.ID
		// 创建规格值
		for _, value := range group.Values {
			specValue := &mod.SpecValue{
				SpecGroupID: specGroup.ID,
				Value:       value,
			}
			if err := curd.SpecValue.WithContext(ctx).Create(specValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReplaceSpecGroups 替换规格组（按名称匹配复用旧ID，只处理变动的部分）
func (r *specGroupRepo) ReplaceSpecGroups(ctx context.Context, productID int64, groups []*biz.SpecGroup) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		return replaceSpecGroupsTx(ctx, tx, productID, groups)
	})
}

// replaceSpecGroupsTx 在指定事务内替换规格组（按名称匹配复用旧ID，只处理变动的部分）
// 供 ReplaceSpecGroups 及商品编辑事务（UpdateProductTx）复用
func replaceSpecGroupsTx(ctx context.Context, tx *curd.Query, productID int64, groups []*biz.SpecGroup) error {
	// 1. 加载旧规格组
	oldGroups, err := tx.SpecGroup.WithContext(ctx).Where(tx.SpecGroup.ProductID.Eq(productID)).Find()
	if err != nil {
		return err
	}
	oldByName := make(map[string]*mod.SpecGroup, len(oldGroups))
	for _, g := range oldGroups {
		oldByName[g.Name] = g
	}

	// 2. 遍历新规格组：匹配的复用旧ID，新增的创建
	for _, group := range groups {
		if old, ok := oldByName[group.Name]; ok {
			// 匹配上：复用旧ID
			group.ID = old.ID
			group.ProductID = productID
			delete(oldByName, group.Name) // 从待删除列表中移除
		} else {
			// 新增的规格组
			specGroup := &mod.SpecGroup{
				ProductID: productID,
				Name:      group.Name,
			}
			if err := tx.SpecGroup.WithContext(ctx).Create(specGroup); err != nil {
				return err
			}
			group.ID = specGroup.ID
			group.ProductID = productID
		}

		// 3. 替换规格值：按值匹配复用，只处理变动
		sgID := group.ID
		oldValues, err := tx.SpecValue.WithContext(ctx).Where(tx.SpecValue.SpecGroupID.Eq(sgID)).Find()
		if err != nil {
			return err
		}
		oldValueSet := make(map[string]struct{}, len(oldValues))
		for _, v := range oldValues {
			oldValueSet[v.Value] = struct{}{}
		}

		// 新增或保留的规格值
		newValueSet := make(map[string]struct{}, len(group.Values))
		for _, value := range group.Values {
			newValueSet[value] = struct{}{}
			if _, exists := oldValueSet[value]; !exists {
				// 新增的规格值
				if err := tx.SpecValue.WithContext(ctx).Create(&mod.SpecValue{
					SpecGroupID: sgID,
					Value:       value,
				}); err != nil {
					return err
				}
			}
		}

		// 删除不再存在的规格值
		for _, v := range oldValues {
			if _, keep := newValueSet[v.Value]; !keep {
				if _, err := tx.SpecValue.WithContext(ctx).Where(
					tx.SpecValue.SpecGroupID.Eq(sgID),
					tx.SpecValue.Value.Eq(v.Value),
				).Delete(); err != nil {
					return err
				}
			}
		}
	}

	// 4. 删除不再存在的旧规格组（及其所有规格值）
	for _, old := range oldByName {
		if _, err := tx.SpecValue.WithContext(ctx).Where(tx.SpecValue.SpecGroupID.Eq(old.ID)).Delete(); err != nil {
			return err
		}
		if _, err := tx.SpecGroup.WithContext(ctx).Where(tx.SpecGroup.ID.Eq(old.ID)).Delete(); err != nil {
			return err
		}
	}

	return nil
}

// GetSpecGroupsByProductID 获取商品的所有规格组
func (r *specGroupRepo) GetSpecGroupsByProductID(ctx context.Context, productID int64) ([]*biz.SpecGroup, error) {
	specGroups, err := curd.SpecGroup.WithContext(ctx).Where(curd.SpecGroup.ProductID.Eq(productID)).Find()
	if err != nil {
		return nil, err
	}
	result := make([]*biz.SpecGroup, 0, len(specGroups))
	for _, sg := range specGroups {
		// 获取规格值
		values, err := curd.SpecValue.WithContext(ctx).Where(curd.SpecValue.SpecGroupID.Eq(sg.ID)).Find()
		if err != nil {
			return nil, err
		}
		valueList := make([]string, 0, len(values))
		for _, v := range values {
			valueList = append(valueList, v.Value)
		}
		result = append(result, &biz.SpecGroup{
			ID:        sg.ID,
			ProductID: sg.ProductID,
			Name:      sg.Name,
			Values:    valueList,
		})
	}
	return result, nil
}

// GetSpecGroupsByProductIDs 批量获取多个商品的规格组，返回 map[productID][]*SpecGroup
func (r *specGroupRepo) GetSpecGroupsByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*biz.SpecGroup, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	// 批量查询规格组
	specGroups, err := curd.SpecGroup.WithContext(ctx).Where(curd.SpecGroup.ProductID.In(productIDs...)).Find()
	if err != nil {
		return nil, err
	}
	if len(specGroups) == 0 {
		return nil, nil
	}
	// 收集所有规格组ID
	groupIDs := make([]int64, 0, len(specGroups))
	for _, sg := range specGroups {
		groupIDs = append(groupIDs, sg.ID)
	}
	// 批量查询规格值
	specValues, err := curd.SpecValue.WithContext(ctx).Where(curd.SpecValue.SpecGroupID.In(groupIDs...)).Find()
	if err != nil {
		return nil, err
	}
	// 按 spec_group_id 分组
	valueMap := make(map[int64][]string)
	for _, sv := range specValues {
		valueMap[sv.SpecGroupID] = append(valueMap[sv.SpecGroupID], sv.Value)
	}
	// 按 productID 分组
	result := make(map[int64][]*biz.SpecGroup)
	for _, sg := range specGroups {
		result[sg.ProductID] = append(result[sg.ProductID], &biz.SpecGroup{
			ID:        sg.ID,
			ProductID: sg.ProductID,
			Name:      sg.Name,
			Values:    valueMap[sg.ID],
		})
	}
	return result, nil
}

// DeleteSpecGroupsByProductID 删除商品的所有规格组
func (r *specGroupRepo) DeleteSpecGroupsByProductID(ctx context.Context, productID int64) error {
	// 先查找所有规格组，删除对应的规格值
	specGroups, err := curd.SpecGroup.WithContext(ctx).Where(curd.SpecGroup.ProductID.Eq(productID)).Find()
	if err != nil {
		return err
	}
	for _, sg := range specGroups {
		if _, err := curd.SpecValue.WithContext(ctx).Where(curd.SpecValue.SpecGroupID.Eq(sg.ID)).Delete(); err != nil {
			return err
		}
	}
	// 删除规格组
	_, err = curd.SpecGroup.WithContext(ctx).Where(curd.SpecGroup.ProductID.Eq(productID)).Delete()
	return err
}

// 辅助函数：将 map[string]string 序列化为 JSON 字符串
func marshalSpec(spec map[string]string) string {
	if len(spec) == 0 {
		return ""
	}
	b, _ := json.Marshal(spec)
	return string(b)
}

// 辅助函数：将 JSON 字符串反序列化为 map[string]string
func unmarshalSpec(specDesc string) map[string]string {
	if specDesc == "" {
		return nil
	}
	spec := make(map[string]string)
	if err := json.Unmarshal([]byte(specDesc), &spec); err != nil {
		return nil
	}
	return spec
}
