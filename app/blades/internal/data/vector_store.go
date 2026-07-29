package data

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	"cyi-app/app/blades/internal/biz"

	"github.com/redis/go-redis/v9"
)

// VectorStore Redis Stack 向量存储
type VectorStore struct {
	client    redis.UniversalClient
	indexName string
	dimension int
}

// NewVectorStore 创建向量存储
func NewVectorStore(client redis.UniversalClient, dimension int) *VectorStore {
	return &VectorStore{
		client:    client,
		indexName: "idx:knowledge",
		dimension: dimension,
	}
}

// CreateIndex 创建向量索引
func (vs *VectorStore) CreateIndex(ctx context.Context) error {
	_, err := vs.client.Do(ctx, "FT.INFO", vs.indexName).Result()
	if err == nil {
		return nil
	}

	cmd := []interface{}{
		"FT.CREATE", vs.indexName, "ON", "HASH", "PREFIX", "1", "kb:",
		"SCHEMA",
		"doc_id", "TAG", "SORTABLE",
		"source", "TAG", "SORTABLE",
		"title", "TEXT", "SORTABLE",
		"content", "TEXT",
		"tags", "TEXT",
		"embedding", "VECTOR", "FLAT", "6",
		"TYPE", "FLOAT32",
		"DIM", strconv.Itoa(vs.dimension),
		"DISTANCE_METRIC", "COSINE",
	}
	_, err = vs.client.Do(ctx, cmd...).Result()
	return err
}

// Store 存储文档块
func (vs *VectorStore) Store(ctx context.Context, doc *biz.KnowledgeDoc) error {
	key := fmt.Sprintf("kb:%s:%d", doc.DocID, doc.ChunkIdx)

	vecBytes := make([]byte, len(doc.Embedding)*4)
	for i, v := range doc.Embedding {
		bits := math.Float32bits(v)
		vecBytes[i*4] = byte(bits)
		vecBytes[i*4+1] = byte(bits >> 8)
		vecBytes[i*4+2] = byte(bits >> 16)
		vecBytes[i*4+3] = byte(bits >> 24)
	}

	tagsJSON, _ := json.Marshal(doc.Tags)

	_, err := vs.client.HSet(ctx, key, map[string]interface{}{
		"doc_id":    doc.DocID,
		"source":    doc.Source,
		"title":     doc.Title,
		"chunk_idx": doc.ChunkIdx,
		"content":   doc.Content,
		"tags":      string(tagsJSON),
		"embedding": vecBytes,
	}).Result()
	return err
}

// Search 向量搜索
// 使用 SCAN + HGETALL 遍历所有文档块，在 Go 端计算余弦相似度，避免 FT.SEARCH 返回格式兼容问题
func (vs *VectorStore) Search(ctx context.Context, source string, queryVec []float32, topK int) ([]*biz.SearchResult, error) {
	type candidate struct {
		sr    *biz.SearchResult
		score float64
	}
	candidates := make([]candidate, 0)

	// 遍历所有 kb: 前缀的 key
	iter := vs.client.Scan(ctx, 0, "kb:*", 200).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		fields, err := vs.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		docSource := fields["source"]
		// source 过滤：匹配指定 source 或 public
		if source != "" && docSource != source && docSource != "public" {
			continue
		}

		// 读取存储的 embedding（二进制字符串）
		embStr := fields["embedding"]
		if embStr == "" {
			continue
		}

		// 将二进制数据还原为 float32 向量
		docVec := bytesToFloat32([]byte(embStr))
		if len(docVec) != len(queryVec) {
			continue
		}

		// 计算余弦相似度
		score := cosineSimilarity(queryVec, docVec)

		tags := make(map[string]string)
		tagsStr := fields["tags"]
		if tagsStr != "" && tagsStr != "null" {
			json.Unmarshal([]byte(tagsStr), &tags)
		}

		candidates = append(candidates, candidate{
			sr: &biz.SearchResult{
				DocID:   fields["doc_id"],
				Title:   fields["title"],
				Content: fields["content"],
				Source:  docSource,
				Score:   score,
				Tags:    tags,
			},
			score: score,
		})
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan for search failed: %w", err)
	}

	// 按相似度降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// 取 Top-K
	if topK > len(candidates) {
		topK = len(candidates)
	}
	results := make([]*biz.SearchResult, 0, topK)
	for i := 0; i < topK; i++ {
		results = append(results, candidates[i].sr)
	}
	return results, nil
}

// bytesToFloat32 将小端字节序列还原为 float32 切片
func bytesToFloat32(b []byte) []float32 {
	count := len(b) / 4
	vec := make([]float32, count)
	for i := 0; i < count; i++ {
		bits := uint32(b[i*4]) | uint32(b[i*4+1])<<8 | uint32(b[i*4+2])<<16 | uint32(b[i*4+3])<<24
		vec[i] = math.Float32frombits(bits)
	}
	return vec
}

// cosineSimilarity 计算两个向量的余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// DeleteDoc 删除文档的所有块
func (vs *VectorStore) DeleteDoc(ctx context.Context, docID string, source string) error {
	pattern := fmt.Sprintf("kb:%s:*", docID)
	iter := vs.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		vs.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

// ListDocuments 列出所有知识文档
// 使用 SCAN + HGETALL 方式遍历 kb: 前缀的 key，避免 FT.SEARCH 返回格式兼容问题
func (vs *VectorStore) ListDocuments(ctx context.Context, source string) ([]*biz.DocInfo, error) {
	// 按 doc_id 去重，统计 chunk 数量
	docMap := make(map[string]*biz.DocInfo)
	chunkCount := make(map[string]int)

	// 使用 SCAN 遍历所有 kb: 前缀的 key
	iter := vs.client.Scan(ctx, 0, "kb:*", 200).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 用 HGETALL 获取该文档块的所有字段
		fields, err := vs.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		docID := fields["doc_id"]
		if docID == "" {
			continue
		}

		docSource := fields["source"]
		// 如果指定了 source 过滤，跳过不匹配的
		if source != "" && docSource != source {
			continue
		}

		if _, exists := docMap[docID]; !exists {
			tags := make(map[string]string)
			tagsStr := fields["tags"]
			if tagsStr != "" && tagsStr != "null" {
				json.Unmarshal([]byte(tagsStr), &tags)
			}
			docMap[docID] = &biz.DocInfo{
				DocID:  docID,
				Source: docSource,
				Title:  fields["title"],
				Tags:   tags,
			}
			chunkCount[docID] = 0
		}
		chunkCount[docID]++
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan documents failed: %w", err)
	}

	// 转换为列表
	docs := make([]*biz.DocInfo, 0, len(docMap))
	for docID, info := range docMap {
		info.ChunkCount = chunkCount[docID]
		docs = append(docs, info)
	}
	return docs, nil
}

// parseFTSearchResult 解析 FT.SEARCH 的返回结果
// 兼容 go-redis v9 多种返回格式：[]interface{} 或 map[interface{}]interface{}
func parseFTSearchResult(result interface{}) []map[string]interface{} {
	entries := make([]map[string]interface{}, 0)

	switch v := result.(type) {
	case []interface{}:
		// 标准格式：[total, key1, fields1[], key2, fields2[], ...]
		if len(v) < 2 {
			return nil
		}
		// v[0] 是总数（int64），后面是 [key, fields, key, fields, ...]
		i := 1
		for i < len(v) {
			// key（string 类型）
			if _, ok := v[i].(string); !ok {
				i++
				continue
			}
			i++
			if i >= len(v) {
				break
			}
			// fields：可能是 []interface{} 或 map[interface{}]interface{}
			entry := convertToFieldMap(v[i])
			if entry != nil {
				entries = append(entries, entry)
			}
			i++
		}

	case map[interface{}]interface{}:
		// go-redis v9 某些版本返回 map 格式
		// 递归遍历，提取所有嵌套的文档字段
		extractEntriesFromMap(v, &entries)

	default:
		// 尝试递归遍历未知类型
		fmt.Printf("[DEBUG parseFTSearchResult] unknown type: %T\n", result)
	}

	return entries
}

// convertToFieldMap 将字段数据（[]interface{} 或 map[interface{}]interface{}）转为 map[string]interface{}
func convertToFieldMap(data interface{}) map[string]interface{} {
	entry := make(map[string]interface{})

	switch v := data.(type) {
	case []interface{}:
		// [field1, value1, field2, value2, ...]
		for j := 0; j < len(v)-1; j += 2 {
			fname, _ := v[j].(string)
			entry[fname] = v[j+1]
		}
	case map[interface{}]interface{}:
		for k, val := range v {
			keyStr, _ := k.(string)
			entry[keyStr] = val
		}
	default:
		return nil
	}

	if len(entry) == 0 {
		return nil
	}
	return entry
}

// extractEntriesFromMap 递归从 map 中提取文档字段
func extractEntriesFromMap(m map[interface{}]interface{}, entries *[]map[string]interface{}) {
	for _, val := range m {
		switch v := val.(type) {
		case map[interface{}]interface{}:
			// 检查是否是文档字段（包含 doc_id 或 title 等）
			entry := make(map[string]interface{})
			for k, fv := range v {
				keyStr, _ := k.(string)
				entry[keyStr] = fv
			}
			if _, ok := entry["doc_id"]; ok {
				*entries = append(*entries, entry)
			} else {
				// 递归查找
				extractEntriesFromMap(v, entries)
			}
		case []interface{}:
			// 可能是 [field, value, ...] 格式
			entry := convertToFieldMap(v)
			if entry != nil {
				if _, ok := entry["doc_id"]; ok {
					*entries = append(*entries, entry)
				}
			}
		}
	}
}
