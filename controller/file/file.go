package file

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/model"
	"GopherAI/service/file"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	UploadFileResponse struct {
		DocumentID string `json:"document_id,omitempty"`
		FilePath   string `json:"file_path,omitempty"`
		controller.Response
	}

	RAGDocumentSummary struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		FileName  string `json:"file_name"`
		FilePath  string `json:"file_path"`
		Source    string `json:"source"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	ListRagDocumentsResponse struct {
		Documents []RAGDocumentSummary `json:"documents"`
		Total     int64                `json:"total"`
		Page      int                  `json:"page"`
		PageSize  int                  `json:"page_size"`
		controller.Response
	}

	GetRagDocumentResponse struct {
		Document *model.RAGDocument `json:"document,omitempty"`
		controller.Response
	}

	DeleteRagDocumentResponse struct {
		controller.Response
	}
)

func UploadRagFile(c *gin.Context) {
	res := new(UploadFileResponse)

	uploadedFile, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
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

func ListRagDocuments(c *gin.Context) {
	res := new(ListRagDocumentsResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("page_size"), 10)
	source := c.Query("source")

	documents, total, err := file.ListRagDocumentsPaged(username, page, pageSize, source)
	if err != nil {
		log.Println("ListRagDocuments fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.Documents = make([]RAGDocumentSummary, 0, len(documents))
	for _, document := range documents {
		res.Documents = append(res.Documents, toRAGDocumentSummary(document))
	}
	res.Total = total
	res.Page = page
	res.PageSize = pageSize
	c.JSON(http.StatusOK, res)
}

func GetRagDocument(c *gin.Context) {
	res := new(GetRagDocumentResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	document, err := file.GetRagDocument(username, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeRecordNotFound))
			return
		}
		log.Println("GetRagDocument fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.Document = document
	c.JSON(http.StatusOK, res)
}

func DeleteRagDocument(c *gin.Context) {
	res := new(DeleteRagDocumentResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	if err := file.DeleteRagDocument(username, documentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, res.CodeOf(code.CodeRecordNotFound))
			return
		}
		log.Println("DeleteRagDocument fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	c.JSON(http.StatusOK, res)
}

func toRAGDocumentSummary(document model.RAGDocument) RAGDocumentSummary {
	return RAGDocumentSummary{
		ID:        document.ID,
		Title:     document.Title,
		FileName:  document.FileName,
		FilePath:  document.FilePath,
		Source:    document.Source,
		CreatedAt: document.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: document.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
