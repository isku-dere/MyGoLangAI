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

	SaveOCRNoteRequest struct {
		Title    string `json:"title" binding:"required"`
		Markdown string `json:"markdown" binding:"required"`
	}

	SaveOCRNoteResponse struct {
		Title      string `json:"title,omitempty"`
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

func StreamSummarizeOCRNotes(c *gin.Context) {
	username := c.GetString("userName")
	if username == "" {
		c.SSEvent("summary_error", gin.H{"message": code.CodeNotLogin.Msg()})
		return
	}

	req := new(SummarizeOCRNotesRequest)
	if err := c.ShouldBindJSON(req); err != nil || len(req.Notes) == 0 {
		c.SSEvent("summary_error", gin.H{"message": code.CodeInvalidParams.Msg()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	result, err := notesservice.StreamOCRNotes(ctx, req.Title, req.Notes, func(delta string) {
		c.SSEvent("summary_delta", gin.H{"delta": delta})
		c.Writer.Flush()
	})
	if err != nil {
		log.Println("StreamSummarizeOCRNotes fail ", err)
		c.SSEvent("summary_error", gin.H{"message": code.AIModelFail.Msg()})
		c.Writer.Flush()
		return
	}

	c.SSEvent("summary_done", gin.H{"title": result.Title, "markdown": result.Markdown})
	c.Writer.Flush()
}

func SaveOCRNote(c *gin.Context) {
	res := new(SaveOCRNoteResponse)
	username := c.GetString("userName")
	if username == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeNotLogin))
		return
	}

	req := new(SaveOCRNoteRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	result, err := notesservice.SaveMarkdownNote(ctx, username, req.Title, req.Markdown)
	if err != nil {
		log.Println("SaveOCRNote fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.Title = result.Title
	res.DocumentID = result.DocumentID
	c.JSON(http.StatusOK, res)
}
