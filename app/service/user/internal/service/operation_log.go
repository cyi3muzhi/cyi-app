package service

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// OperationLogService 操作日志服务（管理端查询）
type OperationLogService struct {
	v1.UnimplementedOperationLogServer
	uc  *biz.OperationLogUseCase
	log log.Logger
}

// NewOperationLogService 创建操作日志服务
func NewOperationLogService(uc *biz.OperationLogUseCase, logger log.Logger) *OperationLogService {
	return &OperationLogService{uc: uc, log: logger}
}

// ListOperationLog 分页查询操作日志
func (s *OperationLogService) ListOperationLog(ctx context.Context, req *v1.ListOperationLogRequest) (*v1.ListOperationLogReply, error) {
	items, total, err := s.uc.List(ctx, req.GetPage(), req.GetPageSize(),
		req.GetStartDate(), req.GetEndDate(), req.GetOperator(), req.GetModule(), req.GetKeyword())
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.OperationLogItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &v1.OperationLogItem{
			Id:        item.ID,
			Operator:  item.Operator,
			Role:      item.Role,
			Module:    item.Module,
			Action:    item.Action,
			Content:   item.Content,
			Method:    item.Method,
			Path:      item.Path,
			Ip:        item.IP,
			UserAgent: item.UserAgent,
			Result:    item.Result,
			CreatedAt: item.CreatedAt,
			Avatar:    item.Avatar,
		})
	}
	return &v1.ListOperationLogReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// CountOperationLog 统计操作日志数量
func (s *OperationLogService) CountOperationLog(ctx context.Context, req *v1.CountOperationLogRequest) (*v1.CountOperationLogReply, error) {
	stats, err := s.uc.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CountOperationLogReply{
		Today: stats.Today,
		Week:  stats.Week,
		Month: stats.Month,
		Total: stats.Total,
	}, nil
}

// ListOperationLogOperators 获取操作人列表
func (s *OperationLogService) ListOperationLogOperators(ctx context.Context, req *v1.ListOperationLogOperatorsRequest) (*v1.ListOperationLogOperatorsReply, error) {
	operators, err := s.uc.ListOperators(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListOperationLogOperatorsReply{
		Operators: operators,
	}, nil
}
