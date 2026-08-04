package biz

import (
	"context"
	"sort"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

// CartEntry 购物车条目（存储层数据结构）
type CartEntry struct {
	ProductID int64
	SkuID     int64
	Quantity  int32
	Selected  bool
}

// CartProduct 购物车商品（聚合商品信息后的展示数据）
type CartProduct struct {
	ProductID int64
	SkuID     int64
	Name      string
	Price     int64 // 单位：分
	Image     string
	Stock     int32 // 可用库存
	Status    int32 // 商品状态 0-下架 1-上架
	Quantity  int32 // 购买数量
	Selected  bool  // 是否选中
	Sku       string
}

// CartRepo 购物车数据仓库接口
type CartRepo interface {
	AddItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error
	GetItems(ctx context.Context, userID int64) ([]*CartEntry, error)
	UpdateItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error
	DeleteItem(ctx context.Context, userID, productID, skuID int64) error
	ToggleItem(ctx context.Context, userID, productID, skuID int64, selected bool) error
	SelectAll(ctx context.Context, userID int64, selected bool) error
	Clear(ctx context.Context, userID int64) error
}

// CartUseCase 购物车业务逻辑
type CartUseCase struct {
	log     *log.Helper
	repo    CartRepo
	product ProductRepo
}

// NewCartUseCase 创建购物车业务用例
func NewCartUseCase(repo CartRepo, product ProductRepo, logger log.Logger) *CartUseCase {
	return &CartUseCase{
		log:     log.NewHelper(logger),
		repo:    repo,
		product: product,
	}
}

// AddItem 加入购物车（同商品同规格数量由存储层累加）
func (u *CartUseCase) AddItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error {
	return u.repo.AddItem(ctx, userID, productID, skuID, quantity)
}

// GetCart 获取购物车：聚合商品名称/价格/图片/库存等信息
// 同一商品不同 SKU 各自成行；未指定 SKU（skuID=0）时按商品第一个 SKU 展示
func (u *CartUseCase) GetCart(ctx context.Context, userID int64) ([]*CartProduct, error) {
	items, err := u.repo.GetItems(ctx, userID)
	if err != nil {
		return nil, err
	}
	products := make([]*CartProduct, 0, len(items))
	for _, item := range items {
		p, err := u.product.GetProduct(ctx, item.ProductID)
		if err != nil {
			// 商品不存在或已删除时跳过该条目，不阻塞整个购物车
			u.log.Warnf("获取商品信息失败 productID=%d: %v", item.ProductID, err)
			continue
		}
		// 按 SKU 聚合价格/库存/图片/规格描述，未匹配到时回退商品级数据
		sku := findSku(p.Skus, item.SkuID)
		cp := &CartProduct{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Name:      p.Name,
			Price:     p.Price,
			Image:     p.Image,
			Stock:     p.Stock,
			Status:    p.Status,
			Quantity:  item.Quantity,
			Selected:  item.Selected,
			Sku:       "默认规格",
		}
		if sku != nil {
			cp.SkuID = sku.ID
			cp.Price = sku.Price
			cp.Stock = sku.Stock
			if sku.Image != "" {
				cp.Image = sku.Image
			}
			if desc := specDesc(sku.Spec); desc != "" {
				cp.Sku = desc
			}
		}
		products = append(products, cp)
	}
	// 排序保证同一商品的多个 SKU 相邻展示
	sort.SliceStable(products, func(i, j int) bool {
		return products[i].ProductID < products[j].ProductID
	})
	return products, nil
}

// UpdateItem 修改购物车商品数量
func (u *CartUseCase) UpdateItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error {
	return u.repo.UpdateItem(ctx, userID, productID, skuID, quantity)
}

// DeleteItem 删除购物车商品
func (u *CartUseCase) DeleteItem(ctx context.Context, userID, productID, skuID int64) error {
	return u.repo.DeleteItem(ctx, userID, productID, skuID)
}

// ToggleItem 选中/取消选中单个商品
func (u *CartUseCase) ToggleItem(ctx context.Context, userID, productID, skuID int64, selected bool) error {
	return u.repo.ToggleItem(ctx, userID, productID, skuID, selected)
}

// SelectAll 全选/取消全选
func (u *CartUseCase) SelectAll(ctx context.Context, userID int64, selected bool) error {
	return u.repo.SelectAll(ctx, userID, selected)
}

// Clear 清空购物车
func (u *CartUseCase) Clear(ctx context.Context, userID int64) error {
	return u.repo.Clear(ctx, userID)
}

// findSku 按 SKU ID 查找商品 SKU；skuID<=0 或未匹配时返回第一个 SKU（无 SKU 返回 nil）
func findSku(skus []*SkuInfo, skuID int64) *SkuInfo {
	if len(skus) == 0 {
		return nil
	}
	if skuID > 0 {
		for _, s := range skus {
			if s.ID == skuID {
				return s
			}
		}
	}
	return skus[0]
}

// specDesc 将规格 map 转换为可读字符串，如 {"颜色":"红色","尺码":"XL"} → "红色/XL"
// 先对 key 排序再拼接，避免 Go map 遍历顺序随机导致规格顺序跳动
func specDesc(spec map[string]string) string {
	if len(spec) == 0 {
		return ""
	}
	keys := make([]string, 0, len(spec))
	for k := range spec {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]string, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, spec[k])
	}
	return strings.Join(vals, "/")
}
