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

// ReplaceSpecGroups 替换规格组（事务保证原子性：先删后插）
func (r *specGroupRepo) ReplaceSpecGroups(ctx context.Context, productID int64, groups []*biz.SpecGroup) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		// 先查找所有旧规格组，删除对应的规格值
		specGroups, err := tx.SpecGroup.WithContext(ctx).Where(tx.SpecGroup.ProductID.Eq(productID)).Find()
		if err != nil {
			return err
		}
		for _, sg := range specGroups {
			if _, err := tx.SpecValue.WithContext(ctx).Where(tx.SpecValue.SpecGroupID.Eq(sg.ID)).Delete(); err != nil {
				return err
			}
		}
		// 删除旧的规格组
		if _, err := tx.SpecGroup.WithContext(ctx).Where(tx.SpecGroup.ProductID.Eq(productID)).Delete(); err != nil {
			return err
		}
		// 创建新的规格组和值
		for _, group := range groups {
			specGroup := &mod.SpecGroup{
				ProductID: productID,
				Name:      group.Name,
			}
			if err := tx.SpecGroup.WithContext(ctx).Create(specGroup); err != nil {
				return err
			}
			group.ID = specGroup.ID
			group.ProductID = productID
			for _, value := range group.Values {
				specValue := &mod.SpecValue{
					SpecGroupID: specGroup.ID,
					Value:       value,
				}
				if err := tx.SpecValue.WithContext(ctx).Create(specValue); err != nil {
					return err
				}
			}
		}
		return nil
	})
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
