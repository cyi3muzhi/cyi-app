package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// LogsService 操作日志服务
type LogsService struct {
	adminv1.UnimplementedAdminLogsServer
	uc  *biz.OperationLogUseCase
	log log.Logger
}

// NewLogsService 创建操作日志服务
func NewLogsService(uc *biz.OperationLogUseCase, logger log.Logger) *LogsService {
	return &LogsService{uc: uc, log: logger}
}

// ListOperationLogs 分页查询操作日志
func (s *LogsService) ListOperationLogs(ctx context.Context, req *adminv1.ListOperationLogsRequest) (*adminv1.ListOperationLogsReply, error) {
	items, total, err := s.uc.List(ctx, req.GetPage(), req.GetPageSize(),
		req.GetStartDate(), req.GetEndDate(), req.GetOperator(), req.GetModule(), req.GetKeyword())
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.AdminLogItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &adminv1.AdminLogItem{
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
	return &adminv1.ListOperationLogsReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// CountOperationLogs 统计操作日志数量
func (s *LogsService) CountOperationLogs(ctx context.Context, req *adminv1.CountOperationLogsRequest) (*adminv1.CountOperationLogsReply, error) {
	stats, err := s.uc.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &adminv1.CountOperationLogsReply{
		Today: stats.Today,
		Week:  stats.Week,
		Month: stats.Month,
		Total: stats.Total,
	}, nil
}

// ListOperationLogOperators 获取操作人列表
func (s *LogsService) ListOperationLogOperators(ctx context.Context, req *adminv1.ListOperationLogOperatorsRequest) (*adminv1.ListOperationLogOperatorsReply, error) {
	operators, err := s.uc.ListOperators(ctx)
	if err != nil {
		return nil, err
	}
	return &adminv1.ListOperationLogOperatorsReply{
		Operators: operators,
	}, nil
}
