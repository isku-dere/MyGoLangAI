package notes

import (
	"GopherAI/common/rag"
	"GopherAI/config"
	ragdocument "GopherAI/dao/rag_document"
	"GopherAI/model"
	"GopherAI/utils"
	"context"
	"fmt"
	"os"
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
	cleaned := make([]SummaryNoteInput, 0, len(inputs))
	for _, input := range inputs {
		markdown := strings.TrimSpace(input.Markdown)
		if markdown == "" {
			continue
		}
		input.Markdown = markdown
		cleaned = append(cleaned, input)
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("no valid markdown notes provided")
	}

	if strings.TrimSpace(title) == "" {
		title = time.Now().Format("2006-01-02") + " OCR ????"
	}

	summary, err := generateSummary(ctx, title, cleaned)
	if err != nil {
		return nil, err
	}

	documentID := utils.GenerateUUID()
	document := &model.RAGDocument{
		ID:       documentID,
		UserName: username,
		Title:    title,
		FileName: documentID + ".md",
		Source:   SummarySource,
		Content:  summary,
	}
	if _, err := ragdocument.Create(document); err != nil {
		return nil, err
	}

	indexer, err := rag.NewRAGIndexer(username, config.GetConfig().RagModelConfig.RagEmbeddingModel)
	if err != nil {
		_ = ragdocument.DeleteByID(username, documentID)
		return nil, err
	}
	if err := indexer.IndexText(ctx, documentID, document.Content, SummarySource); err != nil {
		_ = ragdocument.DeleteByID(username, documentID)
		return nil, err
	}

	return &SummaryResult{Title: title, Markdown: summary, DocumentID: documentID}, nil
}

func generateSummary(ctx context.Context, title string, inputs []SummaryNoteInput) (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if key == "" || modelName == "" || baseURL == "" {
		return "", fmt.Errorf("AI model config is missing")
	}

	llm, err := openaiModel.NewChatModel(ctx, &openaiModel.ChatModelConfig{BaseURL: baseURL, Model: modelName, APIKey: key})
	if err != nil {
		return "", err
	}

	resp, err := llm.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: "????????????????OCR ?????????????????????????????"},
		{Role: schema.User, Content: buildPrompt(title, inputs)},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Content), nil
}

func buildPrompt(title string, inputs []SummaryNoteInput) string {
	var builder strings.Builder
	builder.WriteString("?????? OCR Markdown ?????????????\n")
	builder.WriteString("???\n")
	builder.WriteString("1. ?????? Markdown?\n")
	builder.WriteString("2. ???????????????????\n")
	builder.WriteString("3. ??????????? OCR ?????????????????\n")
	builder.WriteString("4. ?????????????????????????????\n")
	builder.WriteString("5. ?????????????????????????\n\n")
	builder.WriteString("?????")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	for i, input := range inputs {
		builder.WriteString(fmt.Sprintf("--- OCR ?? %d?%s ---\n", i+1, input.FileName))
		builder.WriteString(input.Markdown)
		builder.WriteString("\n\n")
	}
	return builder.String()
}
