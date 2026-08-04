package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// operationLogRepo 操作日志数据仓库（通过 gRPC 调用 user 服务查询）
type operationLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewOperationLogRepo 创建操作日志数据仓库
func NewOperationLogRepo(data *Data, logger log.Logger) biz.OperationLogRepo {
	return &operationLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// List 分页查询操作日志
func (r *operationLogRepo) List(ctx context.Context, page, pageSize int32, startDate, endDate, operator, module, keyword string) ([]*biz.OperationLogItem, int32, error) {
	reply, err := r.data.olc.ListOperationLog(ctx, &userRpc.ListOperationLogRequest{
		Page:      page,
		PageSize:  pageSize,
		StartDate: startDate,
		EndDate:   endDate,
		Operator:  operator,
		Module:    module,
		Keyword:   keyword,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("获取操作日志失败: %w", err)
	}
	items := make([]*biz.OperationLogItem, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, &biz.OperationLogItem{
			ID:        item.Id,
			Operator:  item.Operator,
			Role:      item.Role,
			Module:    item.Module,
			Action:    item.Action,
			Content:   item.Content,
			Method:    item.Method,
			Path:      item.Path,
			IP:        item.Ip,
			UserAgent: item.UserAgent,
			Result:    item.Result,
			CreatedAt: item.CreatedAt,
			Avatar:    r.data.resolvePresignedURL(ctx, item.Avatar, 3600*24*30),
		})
	}
	return items, reply.Total, nil
}

// Count 统计操作日志数量
func (r *operationLogRepo) Count(ctx context.Context) (*biz.OperationLogStats, error) {
	reply, err := r.data.olc.CountOperationLog(ctx, &userRpc.CountOperationLogRequest{})
	if err != nil {
		return nil, fmt.Errorf("统计操作日志失败: %w", err)
	}
	return &biz.OperationLogStats{
		Today: reply.Today,
		Week:  reply.Week,
		Month: reply.Month,
		Total: reply.Total,
	}, nil
}

// ListOperators 获取操作人列表
func (r *operationLogRepo) ListOperators(ctx context.Context) ([]string, error) {
	reply, err := r.data.olc.ListOperationLogOperators(ctx, &userRpc.ListOperationLogOperatorsRequest{})
	if err != nil {
		return nil, fmt.Errorf("获取操作人列表失败: %w", err)
	}
	return reply.Operators, nil
}

// NewNATSConn 创建 NATS 连接（单例）
// 优先从环境变量 NATS_URL 读取地址（Docker 环境应配置为 nats://nats:4222），默认使用 localhost（本地开发）
// 启动时基础设施容器可能尚未就绪，最多重试 5 次；仍失败则返回 nil 降级（不 panic，避免拖垮整个服务）
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(natsURL)
		if err == nil {
			return nc
		}
		time.Sleep(2 * time.Second)
	}
	// 连接失败不 panic：操作日志降级跳过（OperationLogPublisher.Publish 内部有 nil 保护）
	return nil
}

// OperationLogPublisher 操作日志 NATS 发布器
// 将管理端操作日志事件发布到 admin.operation 主题，由 user 服务订阅落库
type OperationLogPublisher struct {
	conn *nats.Conn
	log  *log.Helper
}

// NewOperationLogPublisher 创建操作日志发布器
func NewOperationLogPublisher(conn *nats.Conn, logger log.Logger) *OperationLogPublisher {
	return &OperationLogPublisher{
		conn: conn,
		log:  log.NewHelper(logger),
	}
}

// Publish 发布一条操作日志事件（异步，失败仅记录日志不影响主流程）
func (p *OperationLogPublisher) Publish(evt *event.AdminOperationLog) {
	if p == nil || p.conn == nil || evt == nil {
		return
	}
	data, err := json.Marshal(evt)
	if err != nil {
		p.log.Errorf("序列化操作日志事件失败: %v", err)
		return
	}
	if err := p.conn.Publish(event.SubjectAdminOperation, data); err != nil {
		p.log.Errorf("发布操作日志事件失败: %v", err)
	}
}
