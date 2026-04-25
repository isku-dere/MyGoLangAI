package model

import (
	"time"

	"gorm.io/gorm"
)

type OCRTask struct {
	ID         string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserName   string         `gorm:"index;not null;type:varchar(50)" json:"username"`
	DocumentID string         `gorm:"index;type:varchar(36)" json:"document_id"`
	Status     string         `gorm:"index;type:varchar(20)" json:"status"`
	FileName   string         `gorm:"type:varchar(255)" json:"file_name"`
	FilePath   string         `gorm:"type:varchar(512)" json:"file_path"`
	FileType   int            `json:"file_type"`
	Result     string         `gorm:"type:longtext" json:"result,omitempty"`
	ErrorMsg   string         `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
