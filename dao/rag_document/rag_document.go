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

func ListByUserNamePaged(username string, page, pageSize int, source string) ([]model.RAGDocument, int64, error) {
	var documents []model.RAGDocument
	var total int64
	query := mysql.DB.Model(&model.RAGDocument{}).Where("user_name = ?", username)
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&documents).Error
	return documents, total, err
}

func DeleteByID(username, id string) error {
	return mysql.DB.Where("user_name = ? AND id = ?", username, id).Delete(&model.RAGDocument{}).Error
}
