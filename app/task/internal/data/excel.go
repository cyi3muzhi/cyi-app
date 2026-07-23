package data

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
)

// ExcelGenerator Excel 生成器，实现 biz.ExcelWriter 接口
// 生成订单 Excel 文件并上传到 OSS（RustFS），返回 objectKey 供任务状态记录。
type ExcelGenerator struct {
	log *log.Helper
	oss *rustfs.RustFSClient
}

// NewExcelGenerator 创建 Excel 生成器
func NewExcelGenerator(oss *rustfs.RustFSClient, logger log.Logger) *ExcelGenerator {
	return &ExcelGenerator{
		log: log.NewHelper(logger),
		oss: oss,
	}
}

// GenerateAndUpload 生成订单 Excel 并上传到 OSS，返回 objectKey
// 字段：订单号、下单时间、用户名、商品信息、订单金额(元)、支付方式、订单状态
func (g *ExcelGenerator) GenerateAndUpload(ctx context.Context, orders []*biz.ExportOrder) (string, error) {
	// 1. 生成 Excel
	data, err := g.generateOrderExcel(orders)
	if err != nil {
		return "", fmt.Errorf("生成Excel失败: %w", err)
	}

	// 2. 上传到 OSS
	objectKey := fmt.Sprintf("admin/export/orders_%s.xlsx", time.Now().Format("20060102_150405"))
	bucketName := string(rustfs.AdminServiceBucket)

	// 确保 bucket 存在
	exists, err := g.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := g.oss.CreateBucket(ctx, bucketName); err != nil {
			return "", fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	// 上传文件
	if err := g.oss.UploadObject(ctx, bucketName, objectKey, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("上传导出文件失败: %w", err)
	}

	g.log.Infof("excel uploaded: object_key=%s, size=%d bytes", objectKey, len(data))
	return objectKey, nil
}

// generateOrderExcel 生成订单 Excel 文件，返回字节流
func (g *ExcelGenerator) generateOrderExcel(orders []*biz.ExportOrder) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "订单列表"
	f.SetSheetName(f.GetSheetName(0), sheet)

	// 表头
	headers := []string{"订单号", "下单时间", "用户名", "商品信息", "订单金额(元)", "支付方式", "订单状态"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	// 表头加粗
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetCellStyle(sheet, "A1", "G1", style)

	// 数据行
	for i, order := range orders {
		row := i + 2
		productDesc := formatProducts(order.Products)
		amountYuan := float64(order.Amount) / 100.0 // 分转元

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), order.OrderNo)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), order.CreatedAt)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), order.UserName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), productDesc)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), amountYuan)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), formatPayment(order.Payment))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), order.StatusText)
	}

	// 列宽自适应
	f.SetColWidth(sheet, "A", "A", 25) // 订单号
	f.SetColWidth(sheet, "B", "B", 20) // 下单时间
	f.SetColWidth(sheet, "C", "C", 15) // 用户名
	f.SetColWidth(sheet, "D", "D", 50) // 商品信息
	f.SetColWidth(sheet, "E", "E", 12) // 金额
	f.SetColWidth(sheet, "F", "F", 12) // 支付方式
	f.SetColWidth(sheet, "G", "G", 12) // 状态

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// formatProducts 格式化商品信息为单元格文本
// 格式：商品A(规格) x2; 商品B(规格) x1
func formatProducts(products []*biz.ExportOrderProduct) string {
	if len(products) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(products))
	for _, p := range products {
		// 规格为空时不显示括号
		if p.Spec != "" {
			parts = append(parts, fmt.Sprintf("%s(%s) x%d", p.Name, p.Spec, p.Qty))
		} else {
			parts = append(parts, fmt.Sprintf("%s x%d", p.Name, p.Qty))
		}
	}
	return strings.Join(parts, "; ")
}

// formatPayment 格式化支付方式显示
func formatPayment(payment string) string {
	if payment == "" {
		return "-"
	}
	return payment
}
