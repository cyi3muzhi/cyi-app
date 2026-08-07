package data

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// cartItemValue 购物车条目在 Redis 中的存储值结构
type cartItemValue struct {
	Quantity int32 `json:"quantity"`
	Selected bool  `json:"selected"`
}

// cartRepo 购物车数据仓库
// 存储结构：Redis Hash，key = cart:{userID}，field = {productID}:{skuID}，value = cartItemValue JSON
type cartRepo struct {
	data *Data
	log  *log.Helper
}

// addItemScript 原子化加购 Lua 脚本
// 服务端一次性完成"读取旧值 -> 数量累加 -> 置为选中 -> 写回"，
// 避免并发加购时读改写互相覆盖导致数量丢失
var addItemScript = redis.NewScript(`
local raw = redis.call('HGET', KEYS[1], ARGV[1])
local qty = 0
local selected = 1
if raw then
  local val = cjson.decode(raw)
  qty = val['quantity']
  if val['selected'] == false then
    selected = 0
  end
end
qty = qty + tonumber(ARGV[2])
local newval = cjson.encode({quantity = qty, selected = (selected == 1)})
redis.call('HSET', KEYS[1], ARGV[1], newval)
return qty
`)

// NewCartRepo 创建购物车数据仓库
func NewCartRepo(data *Data, logger log.Logger) biz.CartRepo {
	return &cartRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// cartKey 生成购物车 Redis key
func cartKey(userID int64) string {
	return "cart:" + strconv.FormatInt(userID, 10)
}

// itemField 生成购物车条目的 field：{productID}:{skuID}
func itemField(productID, skuID int64) string {
	return strconv.FormatInt(productID, 10) + ":" + strconv.FormatInt(skuID, 10)
}

// getItem 读取购物车单条（不存在时返回零值 + 空值标记）
func (r *cartRepo) getItem(ctx context.Context, key, field string) (cartItemValue, error) {
	raw, err := r.data.rc.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return cartItemValue{Selected: true}, nil
		}
		return cartItemValue{}, err
	}
	val := cartItemValue{Selected: true}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &val)
	}
	return val, nil
}

// setItem 写入购物车单条
func (r *cartRepo) setItem(ctx context.Context, key, field string, val cartItemValue) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.data.rc.HSet(ctx, key, field, data).Err()
}

// AddItem 加入购物车：同商品同规格数量累加，并置为选中
// 通过 Lua 脚本原子执行，防止并发加购时数量丢失
func (r *cartRepo) AddItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error {
	key := cartKey(userID)
	field := itemField(productID, skuID)
	_, err := addItemScript.Run(ctx, r.data.rc, []string{key}, field, quantity).Result()
	return err
}

// GetItems 获取购物车全部条目
func (r *cartRepo) GetItems(ctx context.Context, userID int64) ([]*biz.CartEntry, error) {
	vals, err := r.data.rc.HGetAll(ctx, cartKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	items := make([]*biz.CartEntry, 0, len(vals))
	for field, raw := range vals {
		productID, skuID := parseField(field)
		val := cartItemValue{Selected: true}
		if raw != "" {
			_ = json.Unmarshal([]byte(raw), &val)
		}
		items = append(items, &biz.CartEntry{
			ProductID: productID,
			SkuID:     skuID,
			Quantity:  val.Quantity,
			Selected:  val.Selected,
		})
	}
	return items, nil
}

// parseField 解析购物车条目 field（兼容旧数据：仅 productID，无 SKU 维度）
func parseField(field string) (productID, skuID int64) {
	parts := strings.SplitN(field, ":", 2)
	productID, _ = strconv.ParseInt(parts[0], 10, 64)
	if len(parts) > 1 {
		skuID, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return
}

// UpdateItem 修改购物车商品数量
func (r *cartRepo) UpdateItem(ctx context.Context, userID, productID, skuID int64, quantity int32) error {
	key := cartKey(userID)
	field := itemField(productID, skuID)
	val, err := r.getItem(ctx, key, field)
	if err != nil {
		return err
	}
	val.Quantity = quantity
	return r.setItem(ctx, key, field, val)
}

// DeleteItem 删除购物车商品（skuID<=0 时删除该商品全部规格）
func (r *cartRepo) DeleteItem(ctx context.Context, userID, productID, skuID int64) error {
	key := cartKey(userID)
	if skuID > 0 {
		field := itemField(productID, skuID)
		n, err := r.data.rc.HDel(ctx, key, field).Result()
		if err != nil {
			return err
		}
		// 兼容旧格式（field 仅为 productID，无 SKU 维度）的遗留数据
		if n == 0 {
			return r.data.rc.HDel(ctx, key, strconv.FormatInt(productID, 10)).Err()
		}
		return nil
	}
	// 未指定 SKU 时删除该商品所有规格条目
	vals, err := r.data.rc.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}
	for field := range vals {
		pid, _ := parseField(field)
		if pid == productID {
			if err := r.data.rc.HDel(ctx, key, field).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

// ToggleItem 选中/取消选中单个商品（skuID<=0 时对该商品全部规格生效）
func (r *cartRepo) ToggleItem(ctx context.Context, userID, productID, skuID int64, selected bool) error {
	key := cartKey(userID)
	if skuID > 0 {
		val, err := r.getItem(ctx, key, itemField(productID, skuID))
		if err != nil {
			return err
		}
		val.Selected = selected
		return r.setItem(ctx, key, itemField(productID, skuID), val)
	}
	// 未指定 SKU 时对该商品全部规格生效
	vals, err := r.data.rc.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}
	for field, raw := range vals {
		pid, _ := parseField(field)
		if pid != productID {
			continue
		}
		val := cartItemValue{Selected: selected}
		if raw != "" {
			_ = json.Unmarshal([]byte(raw), &val)
		}
		val.Selected = selected
		data, _ := json.Marshal(val)
		if err := r.data.rc.HSet(ctx, key, field, data).Err(); err != nil {
			return err
		}
	}
	return nil
}

// SelectAll 全选/取消全选
func (r *cartRepo) SelectAll(ctx context.Context, userID int64, selected bool) error {
	key := cartKey(userID)
	vals, err := r.data.rc.HGetAll(ctx, key).Result()
	if err != nil {
		return err
	}
	for field, raw := range vals {
		val := cartItemValue{Selected: selected}
		if raw != "" {
			_ = json.Unmarshal([]byte(raw), &val)
		}
		val.Selected = selected
		data, _ := json.Marshal(val)
		if err := r.data.rc.HSet(ctx, key, field, data).Err(); err != nil {
			return err
		}
	}
	return nil
}

// Clear 清空购物车
func (r *cartRepo) Clear(ctx context.Context, userID int64) error {
	return r.data.rc.Del(ctx, cartKey(userID)).Err()
}
