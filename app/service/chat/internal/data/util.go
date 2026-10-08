package data

// normalizePage 规范化分页参数：page 最小为 1，pageSize 默认 10、最大 100，
// 防止非法入参导致 SQL offset 溢出或单页数据量过大
func normalizePage(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
