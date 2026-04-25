package ocr

import (
	"GopherAI/config"
	ocrdao "GopherAI/dao/ocr_task"
	"GopherAI/model"
	"GopherAI/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusSucceeded = "succeeded"
	TaskStatusFailed    = "failed"
)

type layoutParsingResponse struct {
	Result struct {
		LayoutParsingResults []struct {
			Markdown struct {
				Text string `json:"text"`
			} `json:"markdown"`
		} `json:"layoutParsingResults"`
	} `json:"result"`
}

func CreateOCRTask(username string, file *multipart.FileHeader) (*model.OCRTask, error) {
	fileType, err := detectOCRFileType(file.Filename)
	if err != nil {
		return nil, err
	}

	taskID := utils.GenerateUUID()
	userDir := filepath.Join("uploads", username, "ocr")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := taskID + ext
	filePath := filepath.Join(userDir, filename)

	if err := saveUploadedFile(file, filePath); err != nil {
		return nil, err
	}

	task := &model.OCRTask{
		ID:       taskID,
		UserName: username,
		Status:   TaskStatusPending,
		FileName: filename,
		FilePath: filePath,
		FileType: fileType,
	}
	if _, err := ocrdao.Create(task); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	go processTask(task.ID, username)
	return task, nil
}

func GetOCRTask(username, taskID string) (*model.OCRTask, error) {
	return ocrdao.GetByUserNameAndID(username, taskID)
}

func IsTerminalStatus(status string) bool {
	return status == TaskStatusSucceeded || status == TaskStatusFailed
}

func processTask(taskID, username string) {
	task, err := ocrdao.GetByUserNameAndID(username, taskID)
	if err != nil {
		log.Printf("Failed to load OCR task %s: %v", taskID, err)
		return
	}

	task.Status = TaskStatusRunning
	if err := ocrdao.Update(task); err != nil {
		log.Printf("Failed to mark OCR task running %s: %v", taskID, err)
		return
	}

	markdown, err := parseFile(task.FilePath, task.FileType)
	if err != nil {
		markFailed(task, err)
		return
	}
	if strings.TrimSpace(markdown) == "" {
		markFailed(task, fmt.Errorf("OCR result is empty"))
		return
	}

	task.Status = TaskStatusSucceeded
	task.DocumentID = ""
	task.Result = markdown
	task.ErrorMsg = ""
	if err := ocrdao.Update(task); err != nil {
		log.Printf("Failed to mark OCR task succeeded %s: %v", taskID, err)
	}
}

func parseFile(filePath string, fileType int) (string, error) {
	cfg := config.GetConfig()
	if cfg.APIURL == "" || cfg.Token == "" {
		return "", fmt.Errorf("OCR API config is missing")
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"file":                      base64.StdEncoding.EncodeToString(fileBytes),
		"fileType":                  fileType,
		"useDocOrientationClassify": false,
		"useDocUnwarping":           false,
		"useChartRecognition":       false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	timeout := time.Duration(cfg.TimeoutSecond) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodPost, cfg.APIURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed layoutParsingResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}

	parts := make([]string, 0, len(parsed.Result.LayoutParsingResults))
	for _, result := range parsed.Result.LayoutParsingResults {
		text := strings.TrimSpace(result.Markdown.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func detectOCRFileType(filename string) (int, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf":
		return 0, nil
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp":
		return 1, nil
	default:
		return 0, fmt.Errorf("unsupported OCR file type: %s", filepath.Ext(filename))
	}
}

func saveUploadedFile(file *multipart.FileHeader, filePath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func markFailed(task *model.OCRTask, err error) {
	task.Status = TaskStatusFailed
	task.ErrorMsg = err.Error()
	if updateErr := ocrdao.Update(task); updateErr != nil {
		log.Printf("Failed to mark OCR task failed %s: %v", task.ID, updateErr)
	}
}
