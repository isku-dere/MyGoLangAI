package router

import (
	"GopherAI/controller/ocr"

	"github.com/gin-gonic/gin"
)

func OCRRouter(r *gin.RouterGroup) {
	r.POST("/upload", ocr.UploadOCRFile)
	r.GET("/tasks/:id", ocr.GetOCRTask)
	r.GET("/tasks/:id/events", ocr.StreamOCRTaskEvents)
	r.POST("/notes/summarize", ocr.SummarizeOCRNotes)
}
