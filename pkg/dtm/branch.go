package dtm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/dtm-labs/dtm/client/dtmcli"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// TCC 分支回调路径约定：{basePath}/{branch}/{op}
const (
	// TCCBasePath TCC 分支回调基础路径（order 服务示例：/api/dtm/stock/try）
	TCCBasePath = "/api/dtm"
	// OpTry TCC 尝试操作
	OpTry = "try"
	// OpConfirm TCC 确认操作
	OpConfirm = "confirm"
	// OpCancel TCC 取消（补偿）操作
	OpCancel = "cancel"
)

// BranchURL 拼接分支回调完整路径：BranchURL("/api/dtm", "stock", "try") → /api/dtm/stock/try
// 供分支注册与 TCC 编排（dtmcli.CallBranch 的 try/confirm/cancel URL）两端共用，保证路径一致
func BranchURL(basePath, branch, op string) string {
	return strings.TrimRight(basePath, "/") + "/" + branch + "/" + op
}

// BranchFunc DTM 分支处理函数（try/confirm/cancel 共用签名）
// ctx: 请求上下文；barrier: 屏障（幂等/空补偿/悬挂保护）；payload: DTM 下发的分支业务载荷（JSON）
type BranchFunc func(ctx context.Context, barrier *dtmcli.BranchBarrier, payload []byte) error

// TCCBranch 一条 TCC 分支定义：分支名 + try/confirm/cancel 三个操作
type TCCBranch struct {
	// Name 分支名，路由为 {basePath}/{Name}/{try|confirm|cancel}
	Name string
	// Try 尝试操作
	Try BranchFunc
	// Confirm 确认操作
	Confirm BranchFunc
	// Cancel 取消（补偿）操作
	Cancel BranchFunc
}

// RegisterTCCBranches 将 TCC 分支注册到 kratos HTTP 服务器
// 每个分支自动注册 {basePath}/{name}/try|confirm|cancel 三个路由，供 DTM 服务回调
func RegisterTCCBranches(s *khttp.Server, basePath string, branches ...TCCBranch) {
	for _, b := range branches {
		s.HandleFunc(BranchURL(basePath, b.Name, OpTry), BranchHandler(b.Try))
		s.HandleFunc(BranchURL(basePath, b.Name, OpConfirm), BranchHandler(b.Confirm))
		s.HandleFunc(BranchURL(basePath, b.Name, OpCancel), BranchHandler(b.Cancel))
	}
}

// BranchHandler 构造 DTM 分支 HTTP 处理器：
// 解析屏障参数（gid/branch_id/trans_type/op）与业务载荷，调用分支逻辑；
// 业务失败（ErrFailure）返回 409（DTM 不再重试），其他错误返回 500（DTM 自动重试）
func BranchHandler(fn BranchFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}
		barrier, err := dtmcli.BarrierFromQuery(r.URL.Query())
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}
		if err := fn(r.Context(), barrier, payload); err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, dtmcli.ErrFailure) {
				code = http.StatusConflict // 409：业务失败，全局事务直接回滚，不重试
			}
			WriteError(w, code, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// WriteError 写出 DTM 分支错误响应
func WriteError(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
}
