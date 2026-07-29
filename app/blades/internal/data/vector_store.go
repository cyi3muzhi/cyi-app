package data

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
func (vs *VectorStore) Search(ctx context.Context, source string, queryVec []float32, topK int) ([]*biz.SearchResult, error) {
	vecBytes := make([]byte, len(queryVec)*4)
	for i, v := range queryVec {
		bits := math.Float32bits(v)
		vecBytes[i*4] = byte(bits)
		vecBytes[i*4+1] = byte(bits >> 8)
		vecBytes[i*4+2] = byte(bits >> 16)
		vecBytes[i*4+3] = byte(bits >> 24)
	}

	filter := "*"
	if source != "" {
		filter = fmt.Sprintf("@source:{%s|public}", source)
	}

	cmd := []interface{}{
		"FT.SEARCH", vs.indexName,
		fmt.Sprintf("%s=>[KNN %d @embedding $vec AS score]", filter, topK),
		"PARAMS", "2", "vec", vecBytes,
		"RETURN", "6", "doc_id", "title", "content", "score", "tags", "source",
		"SORTBY", "score",
		"DIALECT", "2",
	}

	result, err := vs.client.Do(ctx, cmd...).Result()
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	items, ok := result.([]interface{})
	if !ok || len(items) < 2 {
		return nil, nil
	}

	results := make([]*biz.SearchResult, 0)
	for i := 1; i < len(items); i++ {
		entry, ok := items[i].([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		fields, ok := entry[1].([]interface{})
		if !ok {
			continue
		}
		var sr biz.SearchResult
		for j := 0; j < len(fields)-1; j += 2 {
			fieldName, _ := fields[j].(string)
			fieldVal, _ := fields[j+1].(string)
			switch fieldName {
			case "doc_id":
				sr.DocID = fieldVal
			case "title":
				sr.Title = fieldVal
			case "content":
				sr.Content = fieldVal
			case "score":
				sr.Score, _ = strconv.ParseFloat(fieldVal, 64)
				sr.Score = 1 - sr.Score
			case "tags":
				json.Unmarshal([]byte(fieldVal), &sr.Tags)
			}
		}
		if sr.DocID != "" {
			results = append(results, &sr)
		}
	}
	return results, nil
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
