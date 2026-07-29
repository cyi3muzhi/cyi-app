package biz

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TextSplitterConfig 分块配置
type TextSplitterConfig struct {
	ChunkSize    int     // 目标块大小（tokens），默认 800
	ChunkOverlap float64 // 重叠比例 0.0-1.0，默认 0.15
}

// Chunk 文档块
type Chunk struct {
	Index   int
	Content string
	Title   string
}

// TextSplitter 分层文档分块器
type TextSplitter struct {
	config    TextSplitterConfig
	reTitle   *regexp.Regexp
	reNewline *regexp.Regexp
}

// NewTextSplitter 创建分块器
func NewTextSplitter(cfg TextSplitterConfig) *TextSplitter {
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 800
	}
	if cfg.ChunkOverlap <= 0 || cfg.ChunkOverlap > 0.5 {
		cfg.ChunkOverlap = 0.15
	}
	return &TextSplitter{
		config:    cfg,
		reTitle:   regexp.MustCompile(`^(#{1,6})\s+(.+)$`),
		reNewline: regexp.MustCompile(`\n{2,}`),
	}
}

// Split 执行分层分块
func (ts *TextSplitter) Split(docTitle, content string) []*Chunk {
	sections := ts.splitByHeaders(content)
	if len(sections) == 0 {
		sections = []section{{title: docTitle, content: content}}
	}

	var chunks []*Chunk
	for _, sec := range sections {
		subChunks := ts.splitBySize(sec.content)
		for _, c := range subChunks {
			chunks = append(chunks, &Chunk{
				Index:   len(chunks),
				Content: c,
				Title:   sec.title,
			})
		}
	}

	chunks = ts.applyOverlap(chunks)
	return chunks
}

type section struct {
	title   string
	content string
}

func (ts *TextSplitter) splitByHeaders(content string) []section {
	lines := strings.Split(content, "\n")
	var sections []section
	var curTitle string
	var curContent strings.Builder

	for _, line := range lines {
		matches := ts.reTitle.FindStringSubmatch(line)
		if matches != nil {
			if curContent.Len() > 0 {
				sections = append(sections, section{
					title:   curTitle,
					content: strings.TrimSpace(curContent.String()),
				})
				curContent.Reset()
			}
			curTitle = matches[2]
		} else {
			if curContent.Len() > 0 {
				curContent.WriteString("\n")
			}
			curContent.WriteString(line)
		}
	}
	if curContent.Len() > 0 {
		sections = append(sections, section{
			title:   curTitle,
			content: strings.TrimSpace(curContent.String()),
		})
	}
	return sections
}

func (ts *TextSplitter) splitBySize(text string) []string {
	tokens := estimateTokens(text)
	if tokens <= ts.config.ChunkSize {
		return []string{text}
	}

	paragraphs := ts.reNewline.Split(text, -1)
	if len(paragraphs) > 1 {
		return ts.mergeSegments(paragraphs, "\n\n")
	}

	sentences := regexp.MustCompile(`[。！？.!?\n]`).Split(text, -1)
	if len(sentences) > 1 {
		return ts.mergeSegments(sentences, "。")
	}

	return ts.splitByChars(text)
}

func (ts *TextSplitter) mergeSegments(segments []string, sep string) []string {
	var result []string
	var buf strings.Builder

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		newTokens := estimateTokens(buf.String()) + estimateTokens(seg)
		if buf.Len() > 0 && newTokens > ts.config.ChunkSize {
			result = append(result, strings.TrimSpace(buf.String()))
			buf.Reset()
		}
		if buf.Len() > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(seg)
	}
	if buf.Len() > 0 {
		result = append(result, strings.TrimSpace(buf.String()))
	}
	return result
}

func (ts *TextSplitter) splitByChars(text string) []string {
	runes := []rune(text)
	charsPerChunk := ts.config.ChunkSize * 2
	if len(runes) <= charsPerChunk {
		return []string{text}
	}
	var chunks []string
	for i := 0; i < len(runes); i += charsPerChunk {
		end := i + charsPerChunk
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func (ts *TextSplitter) applyOverlap(chunks []*Chunk) []*Chunk {
	if len(chunks) <= 1 {
		return chunks
	}
	result := make([]*Chunk, 0, len(chunks))
	for i, chunk := range chunks {
		if i > 0 {
			prevContent := chunks[i-1].Content
			prevRunes := []rune(prevContent)
			overlapRunes := int(float64(len(prevRunes)) * ts.config.ChunkOverlap)
			if overlapRunes > len(prevRunes) {
				overlapRunes = len(prevRunes)
			}
			if overlapRunes > 0 {
				overlap := string(prevRunes[len(prevRunes)-overlapRunes:])
				chunk.Content = overlap + "\n" + chunk.Content
			}
		}
		result = append(result, chunk)
	}
	return result
}

// estimateTokens 估算 token 数（中英文混排）
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	runes := utf8.RuneCountInString(text)
	return int(math.Ceil(float64(runes) / 2.0))
}
