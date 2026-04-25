package file

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/service/file"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	UploadFileResponse struct {
		DocumentID string `json:"document_id,omitempty"`
		FilePath   string `json:"file_path,omitempty"`
		controller.Response
	}
)

func UploadRagFile(c *gin.Context) {
	res := new(UploadFileResponse)
	uploadedFile, err := c.FormFile("file")
	if err != nil {
		log.Println("FormFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	username := c.GetString("userName")
	if username == "" {
		log.Println("Username not found in context")
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidToken))
		return
	}

	document, err := file.UploadRagFile(username, uploadedFile)
	if err != nil {
		log.Println("UploadFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.DocumentID = document.ID
	res.FilePath = document.FilePath
	c.JSON(http.StatusOK, res)
}
