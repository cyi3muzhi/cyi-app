# 订单状态机文档

## 状态定义

| 枚举值 | 名称 | 说明 |
|--------|------|------|
| 0 | `OrderStatusUnspecified` | 未指定（零值，无效状态，所有操作被拒绝） |
| 1 | `WaitPay` | 待支付（初始状态） |
| 2 | `Pay` | 已支付 |
| 3 | `WaitShip` | 待发货（与 Pay 共享行为） |
| 4 | `Ship` | 已发货 |
| 5 | `WaitReceive` | 待收货（与 Ship 共享行为） |
| 6 | `Complete` | 已完成 |
| 7 | `Cancel` | 已取消（终态） |
| 8 | `Refunding` | 退款中 |
| 9 | `Refunded` | 已退款（终态） |

## 状态转换矩阵

```
              Pay  Ship  Receive  Complete  Cancel  Refunding  Refunded
Unspecified   ✗    ✗      ✗        ✗        ✗       ✗          ✗
WaitPay       ✓    ✗      ✗        ✗        ✓       ✗          ✗
Pay/WaitShip  ✗    ✓      ✗        ✗        ✓       ✓          ✗
Ship/WaitRec  ✗    ✗      ✓        ✗        ✗       ✓          ✗
Complete      ✗    ✗      ✗        ✗        ✗       ✓          ✗
Cancel        ✗    ✗      ✗        ✗        ✗       ✗          ✗
Refunding     ✗    ✗      ✗        ✗        ✗       ✗          ✓
Refunded      ✗    ✗      ✗        ✗        ✗       ✗          ✗
```

## 架构说明

采用**状态模式（State Pattern）** 实现，核心结构：

```
Order (聚合根)
 ├── Status: OrderStatus        — 当前状态枚举
 ├── statusHandler: OrderStatus — 当前状态处理器
 │
 ├── Pay()      → 委托给 statusHandler.Pay()
 ├── Ship()     → 委托给 statusHandler.Ship()
 ├── ...
 │
 └── SetStatus() → 根据新状态创建对应的处理器

OrderStatus (接口)
 ├── statusBase (基类，提供默认实现 + transitionTo)
 │    ├── WaitPayStatus      — 待支付行为
 │    ├── PayStatus           — 已支付/待发货行为
 │    ├── ShipStatus          — 已发货/待收货行为
 │    ├── CompleteStatus      — 已完成行为
 │    ├── CancelStatus        — 已取消行为（终态，全部拒绝）
 │    ├── RefundingStatus     — 退款中行为
 │    └── RefundedStatus      — 已退款行为（终态，全部拒绝）
```

- **默认拒绝**：`statusBase` 为所有操作提供默认实现，返回"无效操作"错误。各具体处理器只需重写**允许**的方法。
- **统一转换**：`transitionTo()` 方法封装状态转换逻辑，先写数据库（失败则回滚内存），成功后切换处理器。

## 正向流程

```
用户下单 → WaitPay → Pay → Ship/WaitReceive → Complete
         ↘         ↗                    ↗
           Cancel  ←  ─ ─ ─ ─ ─ ─ ─ ─ ┘
```

**退款流程**：

```
Pay/Ship/Complete → Refunding → Refunded
```

## 关键文件说明

| 文件 | 职责 |
|------|------|
| `order_status.go` | `OrderStatus` 接口、`statusBase` 基类、`NewOrderStatusHandler` 工厂、`Order` 聚合根的委托方法 |
| `wait_pay_status.go` | 待支付状态：允许 Pay、Cancel |
| `pay_status.go` | 已支付/待发货状态：允许 Ship、Cancel、Refunding |
| `ship_status.go` | 已发货/待收货状态：允许 Receive、Refunding |
| `complete_status.go` | 已完成状态：允许 Refunding |
| `cancel_status.go` | 已取消状态：全部拒绝（终态） |
| `refunding_status.go` | 退款中状态：允许 Refunded |
| `refunded_status.go` | 已退款状态：全部拒绝（终态） |

## 如何新增状态

1. 在 proto 的 `OrderStatus` 枚举中新增值
2. 新建 `xxx_status.go`，实现 `OrderStatus` 接口（只需重写允许的操作）
3. 在 `NewOrderStatusHandler` 工厂中添加 case 分支
4. 在 `getStatusName` 中添加状态的中文描述
5. 更新本文档的状态转换矩阵
6. 在 `order_status_test.go` 中添加对应的测试用例
