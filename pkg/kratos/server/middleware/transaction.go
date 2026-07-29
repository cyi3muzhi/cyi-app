package middleware

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

// Deprecated: 事务管理中间件，已废弃不再使用。
// 建议使用 curd.Q.Transaction 或 data 层的事务管理替代。
// 该实现存在以下问题：
//   - 使用内存 map 存储事务，多实例环境下不可用
//   - 通过 gRPC metadata 传递事务 ID，耦合通信协议
//   - 事务泄漏风险（panic 或超时场景下未正确清理）
func TransactionMiddleware(db *gorm.DB, transactionList []string) middleware.Middleware {
	return selector.Server(
		func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				// 开始事务
				tx := db.Begin()

				// 生成事务ID
				txID := uuid.NewString()

				// 存储事务映射（本地内存或Redis）
				storeTransaction(txID, tx)

				// 创建 Metadata 并设置事务ID
				md := metadata.New(map[string]string{
					"tx-id": txID, // 使用小写键名，gRPC 推荐
				})
				// 将 Metadata 注入上下文
				ctx = metadata.NewOutgoingContext(ctx, md)

				// 延迟清理事务映射
				defer func() {
					if r := recover(); r != nil {
						tx.Rollback()
						removeTransaction(txID)
					}
				}()

				// 执行处理函数
				resp, err := handler(ctx, req)

				// 根据执行结果处理事务
				if err != nil {
					tx.Rollback()
					removeTransaction(txID)
					return nil, err
				}

				// 提交事务
				if err := tx.Commit().Error; err != nil {
					removeTransaction(txID)
					return nil, err
				}

				removeTransaction(txID)
				return resp, nil
			}
		},
	).
		Match(NewTransactionListMatcher(transactionList)).
		Build()
}

func NewTransactionListMatcher(tl []string) selector.MatchFunc {
	transactionListMap := make(map[string]struct{})
	if tl == nil {
		return func(ctx context.Context, operation string) bool {
			return false
		}
	}
	for _, op := range tl {
		transactionListMap[op] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		if _, ok := transactionListMap[operation]; ok {
			return true
		}
		return false
	}
}

// 事务存储（简单实现，生产环境建议使用Redis）
var txMap = make(map[string]*gorm.DB)
var txMu sync.RWMutex

func storeTransaction(txID string, tx *gorm.DB) {
	txMu.Lock()
	defer txMu.Unlock()
	txMap[txID] = tx
}

func getTransaction(txID string) (*gorm.DB, bool) {
	txMu.RLock()
	defer txMu.RUnlock()
	tx, ok := txMap[txID]
	return tx, ok
}

func removeTransaction(txID string) {
	txMu.Lock()
	defer txMu.Unlock()
	delete(txMap, txID)
}
