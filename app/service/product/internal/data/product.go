package data

import (
	"context"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gen"
)

type productRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

// CreateProduct 创建商品
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	// 设置新增字段
	product.Category = in.Category
	product.Image = in.Image
	product.Sales = in.Sales
	if err := curd.Product.WithContext(ctx).Create(product); err != nil {
		return err
	}
	in.ID = product.ID
	return nil
}

// UpdateProduct 更新商品
func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	product.ID = in.ID
	product.Category = in.Category
	product.Image = in.Image
	product.Sales = in.Sales
	if _, err := curd.Product.Debug().WithContext(ctx).Where(curd.Product.ID.Eq(in.ID)).Updates(product); err != nil {
		return err
	}
	return nil
}

// DeleteProduct 删除商品
func (r *productRepo) DeleteProduct(ctx context.Context, productID int64) error {
	_, err := curd.Product.WithContext(ctx).Where(curd.Product.ID.Eq(productID)).Delete()
	return err
}

// GetProduct 获取商品详情
func (r *productRepo) GetProduct(ctx context.Context, productID int64) (*biz.Product, error) {
	product, err := curd.Product.WithContext(ctx).Where(curd.Product.ID.Eq(productID)).First()
	if err != nil {
		return nil, err
	}
	out := &biz.Product{}
	convert.Convert(product, out)
	out.ID = product.ID
	out.Category = product.Category
	out.Image = product.Image
	out.Sales = product.Sales
	return out, nil
}

// ListProduct 获取商品列表（分页 + 关键词搜索 + 分类筛选）
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword, category, status string) ([]*biz.Product, int32, error) {
	// 构建查询条件
	conds := make([]gen.Condition, 0)

	// 关键词搜索：匹配名称或描述
	if keyword != "" {
		conds = append(conds, curd.Product.Name.Like("%"+keyword+"%"))
		conds = append(conds, curd.Product.Description.Like("%"+keyword+"%"))
	}

	// 分类筛选
	if category != "" {
		conds = append(conds, curd.Product.Category.Eq(category))
	}

	// 状态筛选（proto 中 status 是 string，数据库中是 int32）
	if status != "" {
		var statusInt int32
		if status == "上架" || status == "1" {
			statusInt = 1
		} else {
			statusInt = 0
		}
		conds = append(conds, curd.Product.Status.Eq(statusInt))
	}

	// 构建查询
	query := curd.Product.WithContext(ctx)
	if len(conds) > 0 {
		// 如果是关键词搜索，需要用 Or 组合名称和描述条件
		if keyword != "" {
			query = query.Where(curd.Product.Name.Like("%"+keyword+"%")).Or(curd.Product.Description.Like("%"+keyword+"%"))
			// 添加其他 AND 条件
			for _, c := range conds[2:] {
				query = query.Where(c)
			}
		} else {
			for _, c := range conds {
				query = query.Where(c)
			}
		}
	}

	// 计算总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	products, err := query.Offset(offset).Limit(limit).Find()
	if err != nil {
		return nil, 0, err
	}

	// 转换为 biz.Product
	items := make([]*biz.Product, 0, len(products))
	for _, p := range products {
		item := &biz.Product{}
		convert.Convert(p, item)
		item.ID = p.ID
		item.Category = p.Category
		item.Image = p.Image
		item.Sales = p.Sales
		item.Status = biz.ProductStatus(p.Status)
		items = append(items, item)
	}

	return items, int32(total), nil
}
