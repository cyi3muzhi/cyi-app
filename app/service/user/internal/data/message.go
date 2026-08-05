package data

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"cyi-app/app/service/user/internal/biz"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
)

// messageRepo 消息数据仓库实现
type messageRepo struct {
	data *Data
	log  *log.Helper
}

// NewMessageRepo 创建消息数据仓库
func NewMessageRepo(data *Data, logger log.Logger) biz.MessageRepo {
	return &messageRepo{data: data, log: log.NewHelper(logger)}
}

// messageScan 消息列表查询结果（含用户已读状态）
type messageScan struct {
	ID            int64
	Title         string
	Content       string
	MsgType       int32
	CreatedBy     int64
	TargetType    int32
	TargetUserIDs json.RawMessage
	Status        int32
	IsRead        int32
	ReadAt        int64
	CreatedAt     int64
}

// visibleCond 用户可见消息条件（启用 + 全体或指定给该用户）
// 参数 ? 为用户ID字符串（JSON_CONTAINS 第二参数需为 JSON 文档）
const visibleCond = "m.status = 1 AND (m.target_type = 0 OR (m.target_type = 1 AND JSON_CONTAINS(m.target_user_ids, ?)))"

// CreateMessage 创建消息
func (r *messageRepo) CreateMessage(ctx context.Context, m *biz.Message) (int64, error) {
	// 将指定用户ID列表序列化为 JSON 存储（无指定用户时存空数组）
	targetIDs := m.TargetUserIDs
	if targetIDs == nil {
		targetIDs = []int64{}
	}
	raw, err := json.Marshal(targetIDs)
	if err != nil {
		r.log.Errorf("marshal target user ids failed: %v", err)
		return 0, err
	}
	row := &mod.Message{
		Title:         m.Title,
		Content:       m.Content,
		MsgType:       m.MsgType,
		TargetType:    m.TargetType,
		TargetUserIDs: raw,
		Status:        1,
		CreatedBy:     m.CreatedBy,
	}
	if err := r.data.db.WithContext(ctx).Create(row).Error; err != nil {
		r.log.Errorf("create message failed: %v", err)
		return 0, err
	}
	return row.ID, nil
}

// AdminListMessage 管理后台消息列表（分页 + 类型/状态/关键字筛选）
func (r *messageRepo) AdminListMessage(ctx context.Context, page, pageSize int32, msgType, status int32, keyword string) ([]*biz.Message, int32, error) {
	db := r.data.db.WithContext(ctx).Model(&mod.Message{})
	if msgType > 0 {
		db = db.Where("msg_type = ?", msgType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		r.log.Errorf("count admin messages failed: %v", err)
		return nil, 0, err
	}

	var rows []mod.Message
	if err := db.Order("id DESC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&rows).Error; err != nil {
		r.log.Errorf("list admin messages failed: %v", err)
		return nil, 0, err
	}

	items := make([]*biz.Message, 0, len(rows))
	for i := range rows {
		items = append(items, r.adminRowToBiz(&rows[i]))
	}
	return items, int32(total), nil
}

// DeleteMessage 下线/删除消息（软删除）
func (r *messageRepo) DeleteMessage(ctx context.Context, id int64) error {
	res := r.data.db.WithContext(ctx).Where("id = ?", id).Delete(&mod.Message{})
	if res.Error != nil {
		r.log.Errorf("delete message failed, id=%d: %v", id, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		return biz.ErrMessageNotFound
	}
	return nil
}

// ListMyMessage 我的消息列表（分页 + 类型筛选，含已读状态）
func (r *messageRepo) ListMyMessage(ctx context.Context, userID int64, page, pageSize, msgType int32) ([]*biz.Message, int32, error) {
	uid := strconv.FormatInt(userID, 10)
	cond := visibleCond
	args := []interface{}{uid}
	if msgType > 0 {
		cond += " AND m.msg_type = ?"
		args = append(args, msgType)
	}

	// 总数
	var total int64
	if err := r.data.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM message m WHERE "+cond, args...).Scan(&total).Error; err != nil {
		r.log.Errorf("count my messages failed, userID=%d: %v", userID, err)
		return nil, 0, err
	}

	// 列表：左连接 user_message 判断已读状态
	query := `SELECT m.id, m.title, m.content, m.msg_type, m.created_by, m.target_type,
	                 m.target_user_ids, m.status,
	                 IFNULL(um.is_read, 0) AS is_read, IFNULL(um.read_at, 0) AS read_at,
	                 m.create_time AS created_at
	          FROM message m
	          LEFT JOIN user_message um ON um.message_id = m.id AND um.user_id = ?
	          WHERE ` + cond + ` ORDER BY m.id DESC LIMIT ? OFFSET ?`
	queryArgs := append([]interface{}{userID}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)

	var rows []messageScan
	if err := r.data.db.WithContext(ctx).Raw(query, queryArgs...).Scan(&rows).Error; err != nil {
		r.log.Errorf("list my messages failed, userID=%d: %v", userID, err)
		return nil, 0, err
	}

	items := make([]*biz.Message, 0, len(rows))
	for i := range rows {
		items = append(items, r.scanToBiz(&rows[i]))
	}
	return items, int32(total), nil
}

// UnreadMessageCount 未读消息数
func (r *messageRepo) UnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	uid := strconv.FormatInt(userID, 10)
	var count int64
	err := r.data.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM message m
		 LEFT JOIN user_message um ON um.message_id = m.id AND um.user_id = ?
		 WHERE `+visibleCond+` AND (um.is_read IS NULL OR um.is_read = 0)`,
		userID, uid).Scan(&count).Error
	if err != nil {
		r.log.Errorf("count unread messages failed, userID=%d: %v", userID, err)
		return 0, err
	}
	return count, nil
}

// MarkMessageRead 标记单条消息已读
// 先校验消息存在且对用户可见，再 upsert 已读记录（依赖唯一索引 user_id+message_id）
func (r *messageRepo) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	uid := strconv.FormatInt(userID, 10)
	var cnt int64
	if err := r.data.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM message m WHERE m.id = ? AND `+visibleCond,
		messageID, uid).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return biz.ErrMessageNotFound
	}

	now := time.Now().Unix()
	err := r.data.db.WithContext(ctx).Exec(
		`INSERT INTO user_message (user_id, message_id, is_read, read_at, create_time, update_time)
		 VALUES (?, ?, 1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE is_read = 1, read_at = ?, update_time = ?`,
		userID, messageID, now, now, now, now, now).Error
	if err != nil {
		r.log.Errorf("mark message read failed, userID=%d, messageID=%d: %v", userID, messageID, err)
		return err
	}
	return nil
}

// MarkAllMessageRead 全部标记已读（将当前可见的全部消息批量 upsert 已读记录）
func (r *messageRepo) MarkAllMessageRead(ctx context.Context, userID int64) error {
	uid := strconv.FormatInt(userID, 10)
	now := time.Now().Unix()
	err := r.data.db.WithContext(ctx).Exec(
		`INSERT INTO user_message (user_id, message_id, is_read, read_at, create_time, update_time)
		 SELECT ?, m.id, 1, ?, ?, ? FROM message m WHERE `+visibleCond+`
		 ON DUPLICATE KEY UPDATE is_read = 1, read_at = ?, update_time = ?`,
		userID, now, now, now, uid, now, now).Error
	if err != nil {
		r.log.Errorf("mark all messages read failed, userID=%d: %v", userID, err)
		return err
	}
	return nil
}

// adminRowToBiz 管理端消息行 -> biz.Message 转换
func (r *messageRepo) adminRowToBiz(m *mod.Message) *biz.Message {
	targetIDs := []int64{}
	if len(m.TargetUserIDs) > 0 {
		_ = json.Unmarshal(m.TargetUserIDs, &targetIDs)
	}
	return &biz.Message{
		ID:            m.ID,
		Title:         m.Title,
		Content:       m.Content,
		MsgType:       m.MsgType,
		CreatedBy:     m.CreatedBy,
		TargetType:    m.TargetType,
		TargetUserIDs: targetIDs,
		Status:        m.Status,
		CreatedAt:     m.CreateTime,
	}
}

// scanToBiz 用户端消息查询结果 -> biz.Message 转换
func (r *messageRepo) scanToBiz(s *messageScan) *biz.Message {
	return &biz.Message{
		ID:         s.ID,
		Title:      s.Title,
		Content:    s.Content,
		MsgType:    s.MsgType,
		IsRead:     s.IsRead,
		CreatedBy:  s.CreatedBy,
		TargetType: s.TargetType,
		Status:     s.Status,
		ReadAt:     s.ReadAt,
		CreatedAt:  s.CreatedAt,
	}
}

// 确保 messageRepo 实现 biz.MessageRepo（编译期断言）
var _ biz.MessageRepo = (*messageRepo)(nil)
