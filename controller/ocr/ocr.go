package ocr

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/model"
	ocrservice "GopherAI/service/ocr"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	UploadOCRResponse struct {
		TaskID string `json:"task_id,omitempty"`
		Status string `json:"status,omitempty"`
		controller.Response
	}

	GetOCRTaskResponse struct {
		Task *model.OCRTask `json:"task,omitempty"`
		controller.Response
	}
)

func UploadOCRFile(c *gin.Context) {
	res := new(UploadOCRResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	uploadedFile, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	task, err := ocrservice.CreateOCRTask(username, uploadedFile)
	if err != nil {
		log.Println("UploadOCRFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.TaskID = task.ID
	res.Status = task.Status
	c.JSON(http.StatusAccepted, res)
}

func GetOCRTask(c *gin.Context) {
	res := new(GetOCRTaskResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	task, err := ocrservice.GetOCRTask(username, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeRecordNotFound))
			return
		}
		log.Println("GetOCRTask fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.Task = task
	c.JSON(http.StatusOK, res)
}

func StreamOCRTaskEvents(c *gin.Context) {
	username := c.GetString("userName")
	if username == "" {
		c.SSEvent("error", gin.H{"message": code.CodeNotLogin.Msg()})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.SSEvent("error", gin.H{"message": code.CodeInvalidParams.Msg()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		task, err := ocrservice.GetOCRTask(username, taskID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.SSEvent("error", gin.H{"message": code.CodeRecordNotFound.Msg()})
			} else {
				c.SSEvent("error", gin.H{"message": code.CodeServerBusy.Msg()})
			}
			c.Writer.Flush()
			return
		}

		c.SSEvent("ocr_task", task)
		c.Writer.Flush()
		if ocrservice.IsTerminalStatus(task.Status) {
			return
		}

		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
