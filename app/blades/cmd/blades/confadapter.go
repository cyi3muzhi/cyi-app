package main

import (
	"cyi-app/app/blades/internal/biz"
	"cyi-app/app/blades/internal/conf"
)

// provideBladesConfig 将 conf.Blades 转换为 biz.BladesConfig
func provideBladesConfig(cfg *conf.Blades) *biz.BladesConfig {
	if cfg == nil {
		return nil
	}
	bc := &biz.BladesConfig{
		ApiKey: cfg.ApiKey,
		Model:  cfg.Model,
	}
	if cfg.Embedding != nil {
		bc.Embedding = &biz.EmbeddingConfigBiz{
			Model:       cfg.Embedding.Model,
			Dimension:   int(cfg.Embedding.Dimension),
			ApiKey:      cfg.Embedding.ApiKey,
			ApiEndpoint: cfg.Embedding.ApiEndpoint,
		}
	}
	if cfg.Rag != nil {
		bc.Rag = &biz.RAGConfigBiz{
			TopK:           int(cfg.Rag.TopK),
			ChunkSize:      int(cfg.Rag.ChunkSize),
			ChunkOverlap:   float64(cfg.Rag.ChunkOverlap),
			AlwaysRetrieve: cfg.Rag.AlwaysRetrieve,
		}
	}
	return bc
}
