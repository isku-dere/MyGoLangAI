package rag

import (
	"GopherAI/common/redis"
	redisPkg "GopherAI/common/redis"
	"GopherAI/config"
	"context"
	"fmt"
	"os"
	"strings"

	embeddingArk "github.com/cloudwego/eino-ext/components/embedding/ark"
	redisIndexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	redisRetriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"
)

type RAGIndexer struct {
	knowledgeBase string
	keyPrefix     string
	embedding     embedding.Embedder
	indexer       *redisIndexer.Indexer
}

type RAGQuery struct {
	embedding embedding.Embedder
	retriever retriever.Retriever
}

// NewRAGIndexer creates a user-scoped knowledge base indexer.
func NewRAGIndexer(username, embeddingModel string) (*RAGIndexer, error) {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")
	cfg := config.GetConfig()
	dimension := cfg.RagModelConfig.RagDimension

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.RagModelConfig.RagBaseUrl,
		APIKey:  apiKey,
		Model:   embeddingModel,
	}

	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	if err := redisPkg.InitRedisIndex(ctx, username, dimension); err != nil {
		return nil, fmt.Errorf("failed to init redis index: %w", err)
	}

	keyPrefix := redis.GenerateIndexNamePrefix(username)
	indexerConfig := &redisIndexer.IndexerConfig{
		Client:    redisPkg.Rdb,
		KeyPrefix: keyPrefix,
		BatchSize: 10,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redisIndexer.Hashes, error) {
			source := ""
			if s, ok := doc.MetaData["source"].(string); ok {
				source = s
			}

			return &redisIndexer.Hashes{
				Key: doc.ID,
				Field2Value: map[string]redisIndexer.FieldValue{
					"content":  {Value: doc.Content, EmbedKey: "vector"},
					"metadata": {Value: source},
				},
			}, nil
		},
	}
	indexerConfig.Embedding = embedder

	idx, err := redisIndexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &RAGIndexer{
		knowledgeBase: username,
		keyPrefix:     keyPrefix,
		embedding:     embedder,
		indexer:       idx,
	}, nil
}

func NormalizeDocumentID(username, documentID string) string {
	prefix := redis.GenerateIndexNamePrefix(username)
	return strings.TrimPrefix(documentID, prefix)
}

func (r *RAGIndexer) IndexText(ctx context.Context, documentID, content, source string) error {
	documentID = NormalizeDocumentID(r.knowledgeBase, documentID)
	doc := &schema.Document{
		ID:      documentID,
		Content: content,
		MetaData: map[string]any{
			"source": source,
		},
	}

	_, err := r.indexer.Store(ctx, []*schema.Document{doc})
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}
	return nil
}

func (r *RAGIndexer) IndexFile(ctx context.Context, documentID, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return r.IndexText(ctx, documentID, string(content), filePath)
}

func DeleteIndex(ctx context.Context, username string) error {
	if err := redisPkg.DeleteRedisIndex(ctx, username); err != nil {
		return fmt.Errorf("failed to delete redis index: %w", err)
	}
	return nil
}

func DeleteDocument(ctx context.Context, username, documentID string) error {
	documentID = NormalizeDocumentID(username, documentID)
	prefix := redis.GenerateIndexNamePrefix(username)
	keys := make([]string, 0, 1)
	exactKey := prefix + documentID
	if exists, err := redisPkg.Rdb.Exists(ctx, exactKey).Result(); err != nil {
		return fmt.Errorf("failed to check redis document key: %w", err)
	} else if exists > 0 {
		keys = append(keys, exactKey)
	}

	iter := redisPkg.Rdb.Scan(ctx, 0, fmt.Sprintf("%s*%s*", prefix, documentID), 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan redis document keys: %w", err)
	}
	if len(keys) == 0 {
		return nil
	}
	if err := redisPkg.Rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete redis document: %w", err)
	}
	return nil
}

func NewRAGQuery(ctx context.Context, username string) (*RAGQuery, error) {
	cfg := config.GetConfig()
	apiKey := os.Getenv("OPENAI_API_KEY")

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.RagModelConfig.RagBaseUrl,
		APIKey:  apiKey,
		Model:   cfg.RagModelConfig.RagEmbeddingModel,
	}
	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	retrieverConfig := &redisRetriever.RetrieverConfig{
		Client:       redisPkg.Rdb,
		Index:        redis.GenerateIndexName(username),
		Dialect:      2,
		ReturnFields: []string{"content", "metadata", "distance"},
		TopK:         5,
		VectorField:  "vector",
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == "content" {
					resp.Content = val
				} else {
					resp.MetaData[field] = val
				}
			}
			return resp, nil
		},
	}
	retrieverConfig.Embedding = embedder

	rtr, err := redisRetriever.NewRetriever(ctx, retrieverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create retriever: %w", err)
	}

	return &RAGQuery{
		embedding: embedder,
		retriever: rtr,
	}, nil
}

func (r *RAGQuery) RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error) {
	docs, err := r.retriever.Retrieve(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}
	return docs, nil
}

func BuildRAGPrompt(query string, docs []*schema.Document) string {
	if len(docs) == 0 {
		return query
	}

	contextText := ""
	for i, doc := range docs {
		contextText += fmt.Sprintf("[Document %d]: %s\n\n", i+1, doc.Content)
	}

	prompt := fmt.Sprintf(`Answer the user's question based on the following reference documents. If the documents do not contain relevant information, say that no relevant information was found. Reply in the same language as the user's question.

Reference documents:
%s

User question: %s

Provide an accurate and complete answer:`, contextText, query)

	return prompt
}
