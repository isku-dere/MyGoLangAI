package model

import (
	"time"

	"gorm.io/gorm"
)

type RAGDocument struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserName  string         `gorm:"index;not null;type:varchar(50)" json:"username"`
	Title     string         `gorm:"type:varchar(255)" json:"title"`
	FileName  string         `gorm:"type:varchar(255)" json:"file_name"`
	FilePath  string         `gorm:"type:varchar(512)" json:"file_path"`
	Source    string         `gorm:"index;type:varchar(50)" json:"source"`
	Content   string         `gorm:"type:longtext" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
