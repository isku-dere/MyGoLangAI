package ocr_task

import (
	"GopherAI/common/mysql"
	"GopherAI/model"
)

func Create(task *model.OCRTask) (*model.OCRTask, error) {
	err := mysql.DB.Create(task).Error
	return task, err
}

func GetByUserNameAndID(username, id string) (*model.OCRTask, error) {
	var task model.OCRTask
	err := mysql.DB.Where("user_name = ? AND id = ?", username, id).First(&task).Error
	return &task, err
}

func Update(task *model.OCRTask) error {
	return mysql.DB.Save(task).Error
}
