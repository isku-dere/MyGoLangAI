package session

import (
	"GopherAI/common/mysql"
	"GopherAI/model"
	"gorm.io/gorm"
)

func GetSessionsByUserName(userName string) ([]model.Session, error) {
	var sessions []model.Session
	err := mysql.DB.Where("user_name = ?", userName).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

func CreateSession(session *model.Session) (*model.Session, error) {
	err := mysql.DB.Create(session).Error
	return session, err
}

func GetSessionByID(sessionID string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("id = ?", sessionID).First(&session).Error
	return &session, err
}

func GetSessionByUserNameAndID(userName, sessionID string) (*model.Session, error) {
	var session model.Session
	err := mysql.DB.Where("user_name = ? AND id = ?", userName, sessionID).First(&session).Error
	return &session, err
}

func UpdateSessionTitle(userName, sessionID, title string) error {
	return mysql.DB.Model(&model.Session{}).Where("user_name = ? AND id = ?", userName, sessionID).Update("title", title).Error
}

func DeleteSession(userName, sessionID string) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_name = ? AND session_id = ?", userName, sessionID).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		return tx.Where("user_name = ? AND id = ?", userName, sessionID).Delete(&model.Session{}).Error
	})
}
