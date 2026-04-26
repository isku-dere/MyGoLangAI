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

func ListByStatuses(statuses []string) ([]model.OCRTask, error) {
	var tasks []model.OCRTask
	err := mysql.DB.Where("status IN ?", statuses).Find(&tasks).Error
	return tasks, err
}

func ResetStatus(fromStatus, toStatus string) error {
	return mysql.DB.Model(&model.OCRTask{}).Where("status = ?", fromStatus).Update("status", toStatus).Error
}

func ClaimPending(username, id, pendingStatus, runningStatus string) (bool, error) {
	result := mysql.DB.Model(&model.OCRTask{}).
		Where("user_name = ? AND id = ? AND status = ?", username, id, pendingStatus).
		Updates(map[string]any{
			"status":    runningStatus,
			"error_msg": "",
		})
	return result.RowsAffected > 0, result.Error
}
