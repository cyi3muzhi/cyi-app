package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	productRpc "cyi-app/api/service/product/v1"
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type productRepo struct {
	log  *log.Helper
	data *Data
}

// NewProductRepo 创建商品管理数据仓库，通过 gRPC 调用商品服务
func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// uploadImage 处理图片：base64 上传 OSS，否则补全 bucket 前缀
func (r *productRepo) uploadImage(ctx context.Context, imagePath string) (string, error) {
	// 已有完整 URL，无需处理
	if strings.HasPrefix(imagePath, "http") || strings.HasPrefix(imagePath, "data:image") {
		if strings.HasPrefix(imagePath, "data:image") {
			return r.uploadBase64ToOSS(ctx, imagePath)
		}
		return imagePath, nil
	}
	// 已经是 bucket/key 格式（包含 "/"），无需处理
	if strings.Count(imagePath, "/") >= 1 && !strings.HasPrefix(imagePath, string(rustfs.AdminServiceBucket)) {
		return fmt.Sprintf("%s/%s", string(rustfs.AdminServiceBucket), imagePath), nil
	}
	return imagePath, nil
}

// uploadBase64ToOSS 将 base64 图片上传到 OSS
func (r *productRepo) uploadBase64ToOSS(ctx context.Context, base64Data string) (string, error) {

	// 提取纯 base64 数据（去掉 data:image/xxx;base64, 前缀）
	parts := strings.SplitN(base64Data, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的 base64 图片数据")
	}

	objectKey := fmt.Sprintf("admin/product/%s.jpg", uuid.New().String())
	bucketName := string(rustfs.AdminServiceBucket)

	exists, err := r.data.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := r.data.oss.CreateBucket(ctx, bucketName); err != nil {
			return "", fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	body := bytes.NewReader(data)
	if err := r.data.oss.UploadObject(ctx, bucketName, objectKey, body); err != nil {
		return "", fmt.Errorf("上传到OSS失败: %w", err)
	}

	// 返回 bucket/key 路径（读取时生成 presigned URL）
	return fmt.Sprintf("%s/%s", bucketName, objectKey), nil
}

// resolveImageURL 解析图片路径，如果是 OSS 路径则生成 presigned URL
func (r *productRepo) resolveImageURL(ctx context.Context, image string) string {
	if image == "" || strings.HasPrefix(image, "http") {
		return image
	}
	// 格式: bucketName/objectKey
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 2 {
		presigned, err := r.data.oss.GetPresignedURL(ctx, parts[0], parts[1], 3600*24*30)
		if err == nil {
			return presigned
		}
	}
	return image
}

// ListProduct 通过 gRPC 调用商品服务获取商品列表
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*biz.Product, int32, error) {
	gRPCReq := &productRpc.ListProductRequest{
		Page:       page,
		PageSize:   pageSize,
		Keyword:    keyword,
		CategoryId: categoryID,
		Status:     status,
	}
	// 只在明确设置了库存筛选时传值（-1 表示未设置）
	if stockEQ >= 0 {
		gRPCReq.StockEq = wrapperspb.Int32(stockEQ)
	}
	reply, err := r.data.pc.ListProduct(ctx, gRPCReq)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Product, 0, len(reply.Items))
	for _, item := range reply.Items {
		p := protoToBizProduct(item)
		p.Image = r.resolveImageURL(ctx, p.Image)
		// 列表页不需要规格组、SKU 详情和描述，减少响应体积
		p.Specs = nil
		p.Skus = nil
		p.Description = ""
		items = append(items, p)
	}
	return items, reply.GetTotal(), nil
}

// CreateProduct 通过 gRPC 调用商品服务创建商品
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) (int64, error) {
	imagePath, err := r.uploadImage(ctx, in.Image)
	if err != nil {
		return 0, err
	}

	reply, err := r.data.pc.CreateProduct(ctx, &productRpc.CreateProductRequest{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		CategoryId:  in.CategoryID,
		Image:       imagePath,
		Images:      in.Images,
		Sales:       in.Sales,
		Specs:       bizToProtoSpecGroups(in.Specs),
		Skus:        bizToProtoSkus(in.Skus),
	})
	if err != nil {
		return 0, err
	}
	return reply.ProductId, nil
}

// UpdateProduct 通过 gRPC 调用商品服务更新商品
func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	imagePath, err := r.uploadImage(ctx, in.Image)
	if err != nil {
		return err
	}

	var statusInt int32
	if in.Status == "上架" || in.Status == "上架中" {
		statusInt = 1
	} else {
		statusInt = 0
	}

	_, err = r.data.pc.UpdateProduct(ctx, &productRpc.UpdateProductRequest{
		ProductId:   in.ID,
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		CategoryId:  in.CategoryID,
		Image:       imagePath,
		Images:      in.Images,
		Sales:       in.Sales,
		Status:      statusInt,
		Specs:       bizToProtoSpecGroups(in.Specs),
		Skus:        bizToProtoSkus(in.Skus),
	})
	return err
}

// DeleteProduct 通过 gRPC 调用商品服务删除商品
func (r *productRepo) DeleteProduct(ctx context.Context, id int64) error {
	_, err := r.data.pc.DeleteProduct(ctx, &productRpc.DeleteProductRequest{
		ProductId: id,
	})
	return err
}

// BatchOperation 批量操作商品（遍历调用 UpdateProduct/DeleteProduct）
// 跨服务批量操作无法保证原子性，失败时已成功的操作不会回滚
func (r *productRepo) BatchOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	var affected int32
	var errs []string
	for _, id := range ids {
		var err error
		switch action {
		case "删除":
			err = r.DeleteProduct(ctx, id)
		case "上架":
			err = r.batchUpdateStatus(ctx, id, "上架")
		case "下架":
			err = r.batchUpdateStatus(ctx, id, "下架")
		default:
			continue
		}
		if err != nil {
			r.log.Errorf("batch operation %s failed for product %d: %v", action, id, err)
			errs = append(errs, fmt.Sprintf("product %d: %v", id, err))
			continue
		}
		affected++
	}
	if len(errs) > 0 {
		return affected, fmt.Errorf("batch %s partial failure (%d/%d succeeded): %s",
			action, affected, len(ids), strings.Join(errs, "; "))
	}
	return affected, nil
}

// batchUpdateStatus 批量更新商品状态，先查再改以避免覆盖其他字段
func (r *productRepo) batchUpdateStatus(ctx context.Context, id int64, status string) error {
	// 先查现有商品
	product, err := r.GetProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("get product %d failed: %w", id, err)
	}
	if product == nil {
		return fmt.Errorf("product %d not found", id)
	}
	// 只改状态，保留其他字段
	product.Status = status
	return r.UpdateProduct(ctx, product)
}

// protoToBizProduct 将 proto ProductItem 转换为 biz Product
func protoToBizProduct(item *productRpc.ProductItem) *biz.Product {
	status := "上架"
	if item.GetStatus() == 0 {
		status = "下架"
	}
	// 从 SKU 汇总总库存
	stock := int32(0)
	for _, sku := range item.GetSkus() {
		stock += sku.GetStock()
	}
	return &biz.Product{
		ID:          item.GetProductId(),
		Name:        item.GetName(),
		Description: item.GetDescription(),
		Price:       item.GetPrice(),
		Stock:       stock,
		Status:      status,
		CategoryID:  item.GetCategoryId(),
		Image:       item.GetImage(),
		Images:      item.GetImages(),
		Sales:       item.GetSales(),
		Specs:       protoToBizSpecGroups(item.GetSpecs()),
		Skus:        protoToBizSkus(item.GetSkus()),
	}
}

// protoToBizSpecGroups 将 proto SpecGroupInfo 列表转换为 biz SpecGroup 列表
func protoToBizSpecGroups(specs []*productRpc.SpecGroupInfo) []*biz.SpecGroup {
	if len(specs) == 0 {
		return nil
	}
	result := make([]*biz.SpecGroup, 0, len(specs))
	for _, s := range specs {
		result = append(result, &biz.SpecGroup{
			ID:     s.GetId(),
			Name:   s.GetName(),
			Values: s.GetValues(),
		})
	}
	return result
}

// protoToBizSkus 将 proto SkuInfo 列表转换为 biz SkuItem 列表
func protoToBizSkus(skus []*productRpc.SkuInfo) []*biz.SkuItem {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*biz.SkuItem, 0, len(skus))
	for _, s := range skus {
		result = append(result, &biz.SkuItem{
			ID:      s.GetId(),
			SkuCode: s.GetSkuCode(),
			Price:   s.GetPrice(),
			Stock:   s.GetStock(),
			Image:   s.GetImage(),
			Spec:    s.GetSpec(),
		})
	}
	return result
}

// bizToProtoSpecGroups 将 biz SpecGroup 列表转换为 proto SpecGroupInfo 列表
func bizToProtoSpecGroups(specs []*biz.SpecGroup) []*productRpc.SpecGroupInfo {
	if len(specs) == 0 {
		return nil
	}
	result := make([]*productRpc.SpecGroupInfo, 0, len(specs))
	for _, s := range specs {
		result = append(result, &productRpc.SpecGroupInfo{
			Id:     s.ID,
			Name:   s.Name,
			Values: s.Values,
		})
	}
	return result
}

// bizToProtoSkus 将 biz SkuItem 列表转换为 proto SkuInfo 列表
func bizToProtoSkus(skus []*biz.SkuItem) []*productRpc.SkuInfo {
	if len(skus) == 0 {
		return nil
	}
	result := make([]*productRpc.SkuInfo, 0, len(skus))
	for _, s := range skus {
		result = append(result, &productRpc.SkuInfo{
			Id:      s.ID,
			SkuCode: s.SkuCode,
			Price:   s.Price,
			Stock:   s.Stock,
			Image:   s.Image,
			Spec:    s.Spec,
		})
	}
	return result
}

// GetProduct 通过 gRPC 调用商品服务获取商品详情
func (r *productRepo) GetProduct(ctx context.Context, id int64) (*biz.Product, error) {
	reply, err := r.data.pc.GetProduct(ctx, &productRpc.GetProductRequest{
		ProductId: id,
	})
	if err != nil {
		return nil, err
	}
	p := protoToBizProduct(reply.GetProduct())
	p.Image = r.resolveImageURL(ctx, p.Image)
	return p, nil
}
