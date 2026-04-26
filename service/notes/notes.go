package notes

import (
	"GopherAI/common/rag"
	"GopherAI/config"
	ragdocument "GopherAI/dao/rag_document"
	"GopherAI/model"
	"GopherAI/utils"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

const SummarySource = "ocr_summary"

type SummaryNoteInput struct {
	FileName string `json:"fileName"`
	Markdown string `json:"markdown"`
	Edited   bool   `json:"edited"`
}

type SummaryResult struct {
	Title      string `json:"title"`
	Markdown   string `json:"markdown"`
	DocumentID string `json:"document_id"`
}

func SummarizeOCRNotes(ctx context.Context, username, title string, inputs []SummaryNoteInput) (*SummaryResult, error) {
	cleaned, title := normalizeSummaryInput(title, inputs)
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("no valid markdown notes provided")
	}
	summary, err := generateSummary(ctx, title, cleaned)
	if err != nil {
		return nil, err
	}
	return &SummaryResult{Title: title, Markdown: summary}, nil
}

func StreamOCRNotes(ctx context.Context, title string, inputs []SummaryNoteInput, onDelta func(string)) (*SummaryResult, error) {
	cleaned, title := normalizeSummaryInput(title, inputs)
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("no valid markdown notes provided")
	}
	summary, err := streamSummary(ctx, title, cleaned, onDelta)
	if err != nil {
		return nil, err
	}
	return &SummaryResult{Title: title, Markdown: summary}, nil
}

func SaveMarkdownNote(ctx context.Context, username, title, markdown string) (*SummaryResult, error) {
	title = strings.TrimSpace(title)
	markdown = strings.TrimSpace(markdown)
	if title == "" {
		title = time.Now().Format("2006-01-02") + " OCR Note"
	}
	if markdown == "" {
		return nil, fmt.Errorf("markdown is empty")
	}

	documentID := utils.GenerateUUID()
	userDir := filepath.Join("uploads", username, "notes")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, err
	}
	filename := documentID + ".md"
	filePath := filepath.Join(userDir, filename)
	if err := os.WriteFile(filePath, []byte(markdown), 0644); err != nil {
		return nil, err
	}

	document := &model.RAGDocument{ID: documentID, UserName: username, Title: title, FileName: filename, FilePath: filePath, Source: SummarySource, Content: markdown}
	if _, err := ragdocument.Create(document); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	indexer, err := rag.NewRAGIndexer(username, config.GetConfig().RagModelConfig.RagEmbeddingModel)
	if err != nil {
		_ = ragdocument.DeleteByID(username, documentID)
		_ = os.Remove(filePath)
		return nil, err
	}
	if err := indexer.IndexText(ctx, documentID, document.Content, SummarySource); err != nil {
		_ = ragdocument.DeleteByID(username, documentID)
		_ = os.Remove(filePath)
		return nil, err
	}
	return &SummaryResult{Title: title, Markdown: markdown, DocumentID: documentID}, nil
}

func normalizeSummaryInput(title string, inputs []SummaryNoteInput) ([]SummaryNoteInput, string) {
	cleaned := make([]SummaryNoteInput, 0, len(inputs))
	for _, input := range inputs {
		markdown := strings.TrimSpace(input.Markdown)
		if markdown == "" {
			continue
		}
		input.Markdown = markdown
		cleaned = append(cleaned, input)
	}
	if strings.TrimSpace(title) == "" {
		title = time.Now().Format("2006-01-02") + " OCR Note"
	}
	return cleaned, strings.TrimSpace(title)
}

func newSummaryModel(ctx context.Context) (*openaiModel.ChatModel, error) {
	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if key == "" || modelName == "" || baseURL == "" {
		return nil, fmt.Errorf("AI model config is missing")
	}
	return openaiModel.NewChatModel(ctx, &openaiModel.ChatModelConfig{BaseURL: baseURL, Model: modelName, APIKey: key})
}

func generateSummary(ctx context.Context, title string, inputs []SummaryNoteInput) (string, error) {
	llm, err := newSummaryModel(ctx)
	if err != nil {
		return "", err
	}
	resp, err := llm.Generate(ctx, []*schema.Message{{Role: schema.System, Content: systemPrompt()}, {Role: schema.User, Content: buildPrompt(title, inputs)}})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Content), nil
}

func streamSummary(ctx context.Context, title string, inputs []SummaryNoteInput, onDelta func(string)) (string, error) {
	llm, err := newSummaryModel(ctx)
	if err != nil {
		return "", err
	}
	stream, err := llm.Stream(ctx, []*schema.Message{{Role: schema.System, Content: systemPrompt()}, {Role: schema.User, Content: buildPrompt(title, inputs)}})
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var full strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if msg.Content == "" {
			continue
		}
		full.WriteString(msg.Content)
		if onDelta != nil {
			onDelta(msg.Content)
		}
	}
	return strings.TrimSpace(full.String()), nil
}

func systemPrompt() string {
	return "You convert rough handwritten-note markdown into clean study notes. Output only the final Markdown note. Do not include explanations, source labels, metadata, correction notes, error analysis, or unrelated information."
}

func buildPrompt(title string, inputs []SummaryNoteInput) string {
	var builder strings.Builder
	builder.WriteString("Create one complete Markdown note from the OCR markdown below.\n")
	builder.WriteString("Requirements:\n")
	builder.WriteString("1. Output only the Markdown note content.\n")
	builder.WriteString("2. Keep only knowledge and learning content; remove OCR metadata, file names, process notes, apologies, and commentary.\n")
	builder.WriteString("3. Preserve the original note structure as much as possible, including headings, lists, formulas, tables, and code blocks.\n")
	builder.WriteString("4. Rewrite meaningless or corrupted fragments into coherent note content when the intended meaning is clear; remove fragments that cannot be recovered; do not mention recognition errors or correction steps.\n")
	builder.WriteString("5. Use the provided title as the main H1 heading unless the source already has a better H1.\n\n")
	builder.WriteString("Title: ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	for i, input := range inputs {
		builder.WriteString(fmt.Sprintf("--- Source %d ---\n", i+1))
		builder.WriteString(input.Markdown)
		builder.WriteString("\n\n")
	}
	return builder.String()
}
