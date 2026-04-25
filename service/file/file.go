package file

import (
	"GopherAI/common/rag"
	"GopherAI/config"
	ragdocument "GopherAI/dao/rag_document"
	"GopherAI/model"
	"GopherAI/utils"
	"context"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

func UploadRagFile(username string, file *multipart.FileHeader) (*model.RAGDocument, error) {
	if err := utils.ValidateFile(file); err != nil {
		log.Printf("File validation failed: %v", err)
		return nil, err
	}

	userDir := filepath.Join("uploads", username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("Failed to create user directory %s: %v", userDir, err)
		return nil, err
	}

	documentID := utils.GenerateUUID()
	ext := filepath.Ext(file.Filename)
	filename := documentID + ext
	filePath := filepath.Join(userDir, filename)

	src, err := file.Open()
	if err != nil {
		log.Printf("Failed to open uploaded file: %v", err)
		return nil, err
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		log.Printf("Failed to read uploaded file: %v", err)
		return nil, err
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		log.Printf("Failed to write destination file %s: %v", filePath, err)
		return nil, err
	}

	document := &model.RAGDocument{
		ID:       documentID,
		UserName: username,
		Title:    file.Filename,
		FileName: filename,
		FilePath: filePath,
		Source:   "upload",
		Content:  string(content),
	}

	if _, err := ragdocument.Create(document); err != nil {
		os.Remove(filePath)
		log.Printf("Failed to create RAG document record: %v", err)
		return nil, err
	}

	indexer, err := rag.NewRAGIndexer(username, config.GetConfig().RagModelConfig.RagEmbeddingModel)
	if err != nil {
		ragdocument.DeleteByID(username, documentID)
		os.Remove(filePath)
		log.Printf("Failed to create RAG indexer: %v", err)
		return nil, err
	}

	if err := indexer.IndexText(context.Background(), documentID, document.Content, filePath); err != nil {
		ragdocument.DeleteByID(username, documentID)
		os.Remove(filePath)
		log.Printf("Failed to index document: %v", err)
		return nil, err
	}

	log.Printf("RAG document uploaded and indexed successfully: %s", documentID)
	return document, nil
}

func ListRagDocuments(username string) ([]model.RAGDocument, error) {
	return ragdocument.ListByUserName(username)
}

func GetRagDocument(username, documentID string) (*model.RAGDocument, error) {
	return ragdocument.GetByUserNameAndID(username, documentID)
}

func DeleteRagDocument(username, documentID string) error {
	document, err := ragdocument.GetByUserNameAndID(username, documentID)
	if err != nil {
		return err
	}

	if err := rag.DeleteDocument(context.Background(), username, documentID); err != nil {
		return err
	}

	if err := ragdocument.DeleteByID(username, documentID); err != nil {
		return err
	}

	if document.FilePath != "" {
		if err := os.Remove(document.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to remove RAG document file %s: %v", document.FilePath, err)
		}
	}

	return nil
}
