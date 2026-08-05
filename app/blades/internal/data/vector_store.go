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

// Store 存储文档块（支持部分失败回滚）
func (vs *VectorStore) Store(ctx context.Context, doc *biz.KnowledgeDoc) error {
	key := fmt.Sprintf("kb:%s:%d", doc.DocID, doc.ChunkIdx)

	vecBytes := float32ToBytes(doc.Embedding)
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

// StoreBatch 批量存储文档块（部分失败时回滚已写入的 chunk）
func (vs *VectorStore) StoreBatch(ctx context.Context, docs []*biz.KnowledgeDoc) error {
	writtenKeys := make([]string, 0, len(docs))
	for _, doc := range docs {
		key := fmt.Sprintf("kb:%s:%d", doc.DocID, doc.ChunkIdx)
		if err := vs.Store(ctx, doc); err != nil {
			// 回滚已写入的 chunk，避免残留不完整索引
			for _, k := range writtenKeys {
				vs.client.Del(ctx, k)
			}
			return fmt.Errorf("store chunk %s failed (rolled back %d chunks): %w", key, len(writtenKeys), err)
		}
		writtenKeys = append(writtenKeys, key)
	}
	return nil
}

// Search 向量搜索（优先使用 FT.SEARCH KNN，失败时回退到 SCAN）
func (vs *VectorStore) Search(ctx context.Context, source string, queryVec []float32, topK int) ([]*biz.SearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	// 优先使用 FT.SEARCH KNN（利用索引，性能好）
	results, err := vs.searchByFT(ctx, source, queryVec, topK)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// FT.SEARCH 失败或无结果时回退到 SCAN 方式
	return vs.searchByScan(ctx, source, queryVec, topK)
}

// searchByFT 使用 FT.SEARCH KNN 进行向量检索
func (vs *VectorStore) searchByFT(ctx context.Context, source string, queryVec []float32, topK int) ([]*biz.SearchResult, error) {
	vecBytes := float32ToBytes(queryVec)

	// 构建 KNN 查询：*=>[KNN $K @embedding $BLOB] 或带 source 过滤
	var query string
	if source != "" {
		query = fmt.Sprintf("@source:{%s|public}=>[KNN $K @embedding $BLOB]", source)
	} else {
		query = "*=>[KNN $K @embedding $BLOB]"
	}

	args := []interface{}{
		"FT.SEARCH", vs.indexName,
		query,
		"RETURN", "6", "doc_id", "title", "content", "source", "tags", "__embedding_score",
		"LIMIT", "0", strconv.Itoa(topK),
		"DIALECT", "2",
		"PARAMS", "4", "K", strconv.Itoa(topK), "BLOB", vecBytes,
	}

	result, err := vs.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("FT.SEARCH failed: %w", err)
	}

	return parseFTSearchResult(result)
}

// parseFTSearchResult 解析 FT.SEARCH 的返回结果
// 格式：[total, key1, fields1[], key2, fields2[], ...]
func parseFTSearchResult(result interface{}) ([]*biz.SearchResult, error) {
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 2 {
		return nil, nil
	}

	// arr[0] 是总数（int64），后面是 [key, fields, key, fields, ...]
	results := make([]*biz.SearchResult, 0)
	for i := 1; i+1 < len(arr); i += 2 {
		// key（string）
		_, ok := arr[i].(string)
		if !ok {
			continue
		}
		// fields：[]interface{}{field1, value1, field2, value2, ...}
		fields, ok := arr[i+1].([]interface{})
		if !ok {
			continue
		}

		fieldMap := make(map[string]string, len(fields)/2)
		for j := 0; j+1 < len(fields); j += 2 {
			fname, _ := fields[j].(string)
			fval, _ := fields[j+1].(string)
			fieldMap[fname] = fval
		}

		tags := make(map[string]string)
		if tagsStr := fieldMap["tags"]; tagsStr != "" && tagsStr != "null" {
			json.Unmarshal([]byte(tagsStr), &tags)
		}

		var score float64
		if scoreStr := fieldMap["__embedding_score"]; scoreStr != "" {
			if v, err := strconv.ParseFloat(scoreStr, 64); err == nil {
				score = v
			}
		}

		results = append(results, &biz.SearchResult{
			DocID:   fieldMap["doc_id"],
			Title:   fieldMap["title"],
			Content: fieldMap["content"],
			Source:  fieldMap["source"],
			Score:   score,
			Tags:    tags,
		})
	}
	return results, nil
}

// searchByScan 使用 SCAN + HGETALL 方式进行向量检索（fallback）
func (vs *VectorStore) searchByScan(ctx context.Context, source string, queryVec []float32, topK int) ([]*biz.SearchResult, error) {
	type candidate struct {
		sr    *biz.SearchResult
		score float64
	}
	candidates := make([]candidate, 0)

	iter := vs.client.Scan(ctx, 0, "kb:*", 200).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		fields, err := vs.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		docSource := fields["source"]
		if source != "" && docSource != source && docSource != "public" {
			continue
		}

		embStr := fields["embedding"]
		if embStr == "" {
			continue
		}

		docVec := bytesToFloat32([]byte(embStr))
		if len(docVec) != len(queryVec) {
			continue
		}

		score := cosineSimilarity(queryVec, docVec)

		tags := make(map[string]string)
		if tagsStr := fields["tags"]; tagsStr != "" && tagsStr != "null" {
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

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if topK > len(candidates) {
		topK = len(candidates)
	}
	results := make([]*biz.SearchResult, 0, topK)
	for i := 0; i < topK; i++ {
		results = append(results, candidates[i].sr)
	}
	return results, nil
}

// float32ToBytes 将 float32 切片转为小端字节序列
func float32ToBytes(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		bits := math.Float32bits(v)
		buf[i*4] = byte(bits)
		buf[i*4+1] = byte(bits >> 8)
		buf[i*4+2] = byte(bits >> 16)
		buf[i*4+3] = byte(bits >> 24)
	}
	return buf
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
func (vs *VectorStore) ListDocuments(ctx context.Context, source string) ([]*biz.DocInfo, error) {
	docMap := make(map[string]*biz.DocInfo)
	chunkCount := make(map[string]int)

	iter := vs.client.Scan(ctx, 0, "kb:*", 200).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		fields, err := vs.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		docID := fields["doc_id"]
		if docID == "" {
			continue
		}

		docSource := fields["source"]
		if source != "" && docSource != source {
			continue
		}

		if _, exists := docMap[docID]; !exists {
			tags := make(map[string]string)
			if tagsStr := fields["tags"]; tagsStr != "" && tagsStr != "null" {
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

	docs := make([]*biz.DocInfo, 0, len(docMap))
	for docID, info := range docMap {
		info.ChunkCount = chunkCount[docID]
		docs = append(docs, info)
	}
	return docs, nil
}
