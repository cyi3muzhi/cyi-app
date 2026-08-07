package data

import (
	"context"
	"errors"
	"time"

	"cyi-app/app/service/user/internal/biz"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type operationLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewOperationLogRepo 创建操作日志仓储
func NewOperationLogRepo(data *Data, logger log.Logger) biz.OperationLogRepo {
	return &operationLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create 写入一条操作日志（NATS 订阅消费时调用）
func (r *operationLogRepo) Create(ctx context.Context, in *biz.OperationLogRecord) error {
	if in == nil {
		return errors.New("operation log record is nil")
	}
	createdAt := in.CreatedAt
	if createdAt == 0 {
		createdAt = time.Now().Unix()
	}
	model := &mod.AdminOperationLog{
		AdminUserID:  in.AdminUserID,
		OperatorName: in.OperatorName,
		RoleName:     in.RoleName,
		Module:       in.Module,
		Action:       in.Action,
		Content:      in.Content,
		Method:       in.Method,
		Path:         in.Path,
		IP:           in.IP,
		UserAgent:    in.UserAgent,
		RequestBody:  in.RequestBody,
		Result:       in.Result,
		StatusCode:   in.StatusCode,
		DurationMs:   in.DurationMs,
	}
	model.CreateTime = createdAt
	model.UpdateTime = createdAt
	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		r.log.Errorf("写入操作日志失败: %v", err)
		return err
	}
	return nil
}

// List 分页查询操作日志（支持日期/操作人/模块/关键词筛选）
func (r *operationLogRepo) List(ctx context.Context, page, pageSize int32, startDate, endDate, operator, module, keyword string) ([]*biz.OperationLogItem, int32, error) {
	db := r.data.db.WithContext(ctx).Model(&mod.AdminOperationLog{})
	db = r.applyFilters(db, startDate, endDate, operator, module, keyword)

	// 查询总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 分页参数归一化：page 最小 1，pageSize 默认 10、上限 100，防止 int32 溢出
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := int(page-1) * int(pageSize)
	var logs []*mod.AdminOperationLog
	if err := db.Order("id DESC").Offset(offset).Limit(int(pageSize)).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	// 批量查询操作人头像（admin_user 表），避免 N+1 查询
	avatarMap := r.loadAdminAvatars(ctx, logs)

	items := make([]*biz.OperationLogItem, 0, len(logs))
	for _, l := range logs {
		items = append(items, &biz.OperationLogItem{
			ID:        l.ID,
			Operator:  l.OperatorName,
			Role:      l.RoleName,
			Module:    l.Module,
			Action:    l.Action,
			Content:   l.Content,
			Method:    l.Method,
			Path:      l.Path,
			IP:        l.IP,
			UserAgent: l.UserAgent,
			Result:    l.Result,
			CreatedAt: l.CreateTime,
			Avatar:    avatarMap[l.AdminUserID],
		})
	}
	return items, int32(total), nil
}

// loadAdminAvatars 批量查询操作人（管理员）头像，返回 admin_user_id -> 头像路径
func (r *operationLogRepo) loadAdminAvatars(ctx context.Context, logs []*mod.AdminOperationLog) map[int64]string {
	avatarMap := make(map[int64]string, len(logs))
	// 收集去重的管理员 ID
	ids := make([]int64, 0, len(logs))
	seen := make(map[int64]bool)
	for _, l := range logs {
		if l.AdminUserID > 0 && !seen[l.AdminUserID] {
			seen[l.AdminUserID] = true
			ids = append(ids, l.AdminUserID)
		}
	}
	if len(ids) == 0 {
		return avatarMap
	}

	var admins []*mod.AdminUser
	if err := r.data.db.WithContext(ctx).
		Select("id, avatar").
		Where("id IN ?", ids).
		Find(&admins).Error; err != nil {
		r.log.Warnf("批量查询操作人头像失败: %v", err)
		return avatarMap
	}
	for _, a := range admins {
		avatarMap[int64(a.ID)] = a.Avatar
	}
	return avatarMap
}

// Count 统计操作日志数量（今日/近7天/近30天/累计）
// 近7天与近30天采用滚动时间窗口（当前时刻往前推），每次统计使用独立查询，
// 避免 GORM 链式 Where 在同一 Statement 上累加条件导致结果互相污染
func (r *operationLogRepo) Count(ctx context.Context) (*biz.OperationLogStats, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	// 今日零点
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	// 近7天：当前时刻往前推 7 天（滚动窗口）
	startWeek := now.AddDate(0, 0, -7).Unix()
	// 近30天：当前时刻往前推 30 天（滚动窗口）
	startMonth := now.AddDate(0, 0, -30).Unix()

	stats := &biz.OperationLogStats{}
	model := &mod.AdminOperationLog{}
	if err := r.data.db.WithContext(ctx).Model(model).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	if err := r.data.db.WithContext(ctx).Model(model).Where("create_time >= ?", startToday).Count(&stats.Today).Error; err != nil {
		return nil, err
	}
	if err := r.data.db.WithContext(ctx).Model(model).Where("create_time >= ?", startWeek).Count(&stats.Week).Error; err != nil {
		return nil, err
	}
	if err := r.data.db.WithContext(ctx).Model(model).Where("create_time >= ?", startMonth).Count(&stats.Month).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// ListOperators 获取操作人列表（按姓名去重）
func (r *operationLogRepo) ListOperators(ctx context.Context) ([]string, error) {
	var operators []string
	if err := r.data.db.WithContext(ctx).Model(&mod.AdminOperationLog{}).
		Distinct("operator_name").
		Where("operator_name != ''").
		Order("operator_name").
		Pluck("operator_name", &operators).Error; err != nil {
		return nil, err
	}
	return operators, nil
}

// applyFilters 组合筛选条件（日期为 2006-01-02 字符串，转 Unix 时间戳区间）
func (r *operationLogRepo) applyFilters(db *gorm.DB, startDate, endDate, operator, module, keyword string) *gorm.DB {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	if startDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", startDate, loc); err == nil {
			db = db.Where("create_time >= ?", t.Unix())
		}
	}
	if endDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", endDate, loc); err == nil {
			db = db.Where("create_time <= ?", t.Add(24*time.Hour-time.Second).Unix())
		}
	}
	if operator != "" {
		db = db.Where("operator_name = ?", operator)
	}
	if module != "" {
		db = db.Where("module = ?", module)
	}
	if keyword != "" {
		db = db.Where("content LIKE ?", "%"+keyword+"%")
	}
	return db
}
