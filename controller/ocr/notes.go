package ocr

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	notesservice "GopherAI/service/notes"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type (
	SummarizeOCRNotesRequest struct {
		Title string                          `json:"title"`
		Notes []notesservice.SummaryNoteInput `json:"notes" binding:"required"`
	}

	SummarizeOCRNotesResponse struct {
		Title      string `json:"title,omitempty"`
		Markdown   string `json:"markdown,omitempty"`
		DocumentID string `json:"document_id,omitempty"`
		controller.Response
	}
)

func SummarizeOCRNotes(c *gin.Context) {
	res := new(SummarizeOCRNotesResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	req := new(SummarizeOCRNotesRequest)
	if err := c.ShouldBindJSON(req); err != nil || len(req.Notes) == 0 {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	result, err := notesservice.SummarizeOCRNotes(ctx, username, req.Title, req.Notes)
	if err != nil {
		log.Println("SummarizeOCRNotes fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.AIModelFail))
		return
	}

	res.Success()
	res.Title = result.Title
	res.Markdown = result.Markdown
	res.DocumentID = result.DocumentID
	c.JSON(http.StatusOK, res)
}
