package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient DeepSeek Embedding API 客户端
type EmbeddingClient struct {
	apiKey      string
	apiEndpoint string
	model       string
	dimension   int
	client      *http.Client
}

// NewEmbeddingClient 创建 Embedding 客户端
func NewEmbeddingClient(apiKey, apiEndpoint, model string, dimension int) *EmbeddingClient {
	if apiEndpoint == "" {
		apiEndpoint = "https://api.deepseek.com/v1"
	}
	if model == "" {
		model = "deepseek-embedding"
	}
	if dimension <= 0 {
		dimension = 768
	}
	return &EmbeddingClient{
		apiKey:      apiKey,
		apiEndpoint: apiEndpoint,
		model:       model,
		dimension:   dimension,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Embed 将文本转为向量
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := c.BatchEmbed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("embedding returned empty result")
	}
	return results[0], nil
}

// BatchEmbed 批量将文本转为向量
func (c *EmbeddingClient) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	apiURL := c.apiEndpoint + "/embeddings"
	requestBody := map[string]interface{}{
		"model": c.model,
		"input": texts,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding api error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	results := make([][]float32, 0, len(apiResp.Data))
	for _, d := range apiResp.Data {
		vec := make([]float32, len(d.Embedding))
		for i, v := range d.Embedding {
			vec[i] = float32(v)
		}
		results = append(results, vec)
	}
	return results, nil
}
