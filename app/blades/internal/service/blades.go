package service

import (
	"context"
	"io"

	pb "cyi-app/api/blades/v1"
	"cyi-app/app/blades/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// BladesService 实现 Blades 服务接口
type BladesService struct {
	pb.UnimplementedBladesServer
	agentUC     *biz.AgentUseCase
	knowledgeUC *biz.KnowledgeUseCase
	log         *log.Helper
}

// NewBladesService 创建 BladesService
func NewBladesService(agentUC *biz.AgentUseCase, knowledgeUC *biz.KnowledgeUseCase, logger log.Logger) *BladesService {
	return &BladesService{
		agentUC:     agentUC,
		knowledgeUC: knowledgeUC,
		log:         log.NewHelper(logger),
	}
}

// Chat 非流式聊天
func (s *BladesService) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatReply, error) {
	response, err := s.agentUC.Chat(ctx, req.UserId, req.Source, req.Role, req.Message, req.EnableRag)
	if err != nil {
		s.log.Error("Chat failed: %v", err)
		return nil, err
	}
	return &pb.ChatReply{Message: response}, nil
}

// ChatStream 流式聊天（服务端流式）
func (s *BladesService) ChatStream(req *pb.ChatStreamRequest, stream pb.Blades_ChatStreamServer) error {
	err := s.agentUC.ChatStream(stream.Context(), req.UserId, req.Source, req.Role, req.Message, req.EnableRag,
		func(chunk string, done bool) {
			resp := &pb.ChatChunkReply{Chunk: chunk, Done: done}
			if err := stream.Send(resp); err != nil && err != io.EOF {
				s.log.Error("ChatStream Send failed: %v", err)
			}
		})
	if err != nil {
		s.log.Error("Agent ChatStream failed: %v", err)
		return err
	}
	return nil
}

// GetChatHistory 获取聊天历史
func (s *BladesService) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.GetChatHistoryReply, error) {
	history, err := s.agentUC.LoadChatHistory(ctx, req.UserId, req.Source)
	if err != nil {
		s.log.Error("GetChatHistory failed: %v", err)
		return nil, err
	}
	messages := make([]*pb.HistoryMessage, 0, len(history))
	for _, m := range history {
		messages = append(messages, &pb.HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	return &pb.GetChatHistoryReply{Messages: messages}, nil
}

// ClearChatHistory 清除聊天历史
func (s *BladesService) ClearChatHistory(ctx context.Context, req *pb.ClearChatHistoryRequest) (*pb.ClearChatHistoryReply, error) {
	err := s.agentUC.ClearHistory(ctx, req.UserId, req.Source)
	if err != nil {
		s.log.Error("ClearChatHistory failed: %v", err)
		return nil, err
	}
	return &pb.ClearChatHistoryReply{Success: true}, nil
}

// UploadKnowledge 上传知识文档
func (s *BladesService) UploadKnowledge(stream pb.Blades_UploadKnowledgeServer) error {
	var docID, source, title, contentType string
	var tags map[string]string
	var content []byte

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch req.Data.(type) {
		case *pb.UploadKnowledgeRequest_Meta:
			meta := req.GetMeta()
			docID = meta.DocId
			source = meta.Source
			title = meta.Title
			contentType = meta.ContentType
			tags = meta.Tags
		case *pb.UploadKnowledgeRequest_ContentChunk:
			content = append(content, req.GetContentChunk()...)
		}
	}

	chunkCount, _, err := s.knowledgeUC.UploadDocument(stream.Context(), docID, source, title, contentType, content, tags)
	if err != nil {
		s.log.Error("UploadKnowledge failed: %v", err)
		return err
	}

	return stream.SendAndClose(&pb.UploadKnowledgeReply{
		DocId:      docID,
		ChunkCount: int32(chunkCount),
	})
}

// SearchKnowledge 搜索知识库
func (s *BladesService) SearchKnowledge(ctx context.Context, req *pb.SearchKnowledgeRequest) (*pb.SearchKnowledgeReply, error) {
	results, err := s.knowledgeUC.SearchKnowledge(ctx, req.Source, req.Query, int(req.TopK))
	if err != nil {
		s.log.Error("SearchKnowledge failed: %v", err)
		return nil, err
	}

	chunks := make([]*pb.KnowledgeChunk, 0)
	for _, r := range results {
		chunks = append(chunks, &pb.KnowledgeChunk{
			DocId:   r.DocID,
			Title:   r.Title,
			Content: r.Content,
			Score:   float32(r.Score),
			Tags:    r.Tags,
		})
	}
	return &pb.SearchKnowledgeReply{Chunks: chunks}, nil
}

// DeleteKnowledge 删除知识文档
func (s *BladesService) DeleteKnowledge(ctx context.Context, req *pb.DeleteKnowledgeRequest) (*pb.DeleteKnowledgeReply, error) {
	err := s.knowledgeUC.DeleteDocument(ctx, req.DocId, req.Source)
	if err != nil {
		s.log.Error("DeleteKnowledge failed: %v", err)
		return nil, err
	}
	return &pb.DeleteKnowledgeReply{Success: true}, nil
}

// ListKnowledge 列出知识库文档
func (s *BladesService) ListKnowledge(ctx context.Context, req *pb.ListKnowledgeRequest) (*pb.ListKnowledgeReply, error) {
	docs, err := s.knowledgeUC.ListDocuments(ctx, req.Source)
	if err != nil {
		s.log.Error("ListKnowledge failed: %v", err)
		return nil, err
	}

	pbDocs := make([]*pb.KnowledgeDocInfo, 0, len(docs))
	for _, d := range docs {
		pbDocs = append(pbDocs, &pb.KnowledgeDocInfo{
			DocId:      d.DocID,
			Source:     d.Source,
			Title:      d.Title,
			ChunkCount: int32(d.ChunkCount),
			Tags:       d.Tags,
		})
	}
	return &pb.ListKnowledgeReply{Docs: pbDocs}, nil
}

// GenerateText 生成文本（不保存聊天历史）
func (s *BladesService) GenerateText(ctx context.Context, req *pb.GenerateTextRequest) (*pb.GenerateTextReply, error) {
	text, err := s.agentUC.GenerateText(ctx, req.Prompt)
	if err != nil {
		s.log.Error("GenerateText failed: %v", err)
		return nil, err
	}
	return &pb.GenerateTextReply{Text: text}, nil
}
