package biz

import (
	"time"
)

// LogisticsTrace 物流轨迹节点
type LogisticsTrace struct {
	Time   int64  // 节点时间（Unix 秒）
	Desc   string // 节点描述
	Status int32  // 节点状态: 1-已发货 2-运输中 3-派送中 4-已签收
}

// Logistics 物流信息（发货后）
type Logistics struct {
	OrderID          int64             // 订单ID
	LogisticsCompany string            // 物流公司
	LogisticsNo      string            // 快递单号
	ShipTime         int64             // 发货时间（Unix 秒，0=未发货）
	Traces           []*LogisticsTrace // 轨迹节点（按时间升序）
}

// traceStage 轨迹阶段定义（相对发货时间的偏移）
type traceStage struct {
	offset time.Duration // 相对发货时间的偏移
	desc   string        // 节点描述
	status int32         // 节点状态
}

// logisticsStages 模拟物流轨迹阶段表
// 当前为演示用模拟物流，按发货时间推算轨迹节点；后续可替换为真实物流 API 查询
var logisticsStages = []traceStage{
	{0 * time.Hour, "商家已发货，等待快递揽收", 1},
	{30 * time.Minute, "快递已揽收，包裹到达始发网点", 1},
	{2 * time.Hour, "包裹已发出，运输中", 2},
	{8 * time.Hour, "包裹到达中转站，正在转运", 2},
	{16 * time.Hour, "包裹到达目的地城市", 2},
	{20 * time.Hour, "派送中，快递员正在派送", 3},
	{26 * time.Hour, "包裹已签收，感谢您的信任", 4},
}

// BuildLogisticsTraces 根据发货时间与当前时间推算物流轨迹节点
// 只返回已到达时间的节点；订单已完成时强制返回全部节点（含签收）
// @param shipTime 发货时间（Unix 秒）
// @param now 当前时间（Unix 秒）
// @param orderComplete 订单是否已完成
// @return 按时间升序的轨迹节点列表
func BuildLogisticsTraces(shipTime, now int64, orderComplete bool) []*LogisticsTrace {
	traces := make([]*LogisticsTrace, 0, len(logisticsStages))
	for _, stage := range logisticsStages {
		nodeTime := shipTime + int64(stage.offset/time.Second)
		// 未到时间的节点不展示，除非订单已完成（强制补全轨迹线）
		if nodeTime > now && !orderComplete {
			continue
		}
		// 订单已完成但签收节点计划时间在未来时，以当前时间作为签收时间
		if nodeTime > now {
			nodeTime = now
		}
		traces = append(traces, &LogisticsTrace{
			Time:   nodeTime,
			Desc:   stage.desc,
			Status: stage.status,
		})
	}
	return traces
}
