package router

import (
	"GopherAI/controller/file"

	"github.com/gin-gonic/gin"
)

func FileRouter(r *gin.RouterGroup) {
	r.POST("/upload", file.UploadRagFile)
	r.GET("/documents", file.ListRagDocuments)
	r.GET("/documents/:id", file.GetRagDocument)
	r.DELETE("/documents/:id", file.DeleteRagDocument)
}
