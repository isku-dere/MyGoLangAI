package rag_document

import (
	"GopherAI/common/mysql"
	"GopherAI/model"
)

func Create(document *model.RAGDocument) (*model.RAGDocument, error) {
	err := mysql.DB.Create(document).Error
	return document, err
}

func GetByID(id string) (*model.RAGDocument, error) {
	var document model.RAGDocument
	err := mysql.DB.Where("id = ?", id).First(&document).Error
	return &document, err
}

func GetByUserNameAndID(username, id string) (*model.RAGDocument, error) {
	var document model.RAGDocument
	err := mysql.DB.Where("user_name = ? AND id = ?", username, id).First(&document).Error
	return &document, err
}

func ListByUserName(username string) ([]model.RAGDocument, error) {
	var documents []model.RAGDocument
	err := mysql.DB.Where("user_name = ?", username).Order("created_at desc").Find(&documents).Error
	return documents, err
}

func DeleteByID(username, id string) error {
	return mysql.DB.Where("user_name = ? AND id = ?", username, id).Delete(&model.RAGDocument{}).Error
}
