package session

import (
	"GopherAI/common/aihelper"
	"GopherAI/common/code"
	messagedao "GopherAI/dao/message"
	sessiondao "GopherAI/dao/session"
	"GopherAI/model"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ctx = context.Background()

func GetUserSessionsByUserName(userName string) ([]model.SessionInfo, error) {
	sessions, err := sessiondao.GetSessionsByUserName(userName)
	if err != nil {
		return nil, err
	}

	infos := make([]model.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			title = "New Chat"
		}
		infos = append(infos, model.SessionInfo{SessionID: s.ID, Title: title})
	}
	return infos, nil
}
func CreateSessionAndSendMessage(userName string, userQuestion string, modelType string) (string, string, code.Code) {
	//1：创建一个新的会话
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: userName,
		Title:    summarizeSessionTitle(userQuestion), // 可以根据需求设置标题，这边暂时用用户第一次的问题作为标题
	}
	createdSession, err := sessiondao.CreateSession(newSession)
	if err != nil {
		log.Println("CreateSessionAndSendMessage CreateSession error:", err)
		return "", "", code.CodeServerBusy
	}

	//2：获取AIHelper并通过其管理消息
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey":   "your-api-key", // TODO: 从配置中获取
		"username": userName,       // 用于 RAG 模型获取用户文档
	}
	helper, err := manager.GetOrCreateAIHelper(userName, createdSession.ID, modelType, config)
	if err != nil {
		log.Println("CreateSessionAndSendMessage GetOrCreateAIHelper error:", err)
		return "", "", code.AIModelFail
	}

	//3：生成AI回复
	aiResponse, err_ := helper.GenerateResponse(userName, ctx, userQuestion)
	if err_ != nil {
		log.Println("CreateSessionAndSendMessage GenerateResponse error:", err_)
		return "", "", code.AIModelFail
	}

	return createdSession.ID, aiResponse.Content, code.CodeSuccess
}

func CreateStreamSessionOnly(userName string, userQuestion string) (string, code.Code) {
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: userName,
		Title:    summarizeSessionTitle(userQuestion),
	}
	createdSession, err := sessiondao.CreateSession(newSession)
	if err != nil {
		log.Println("CreateStreamSessionOnly CreateSession error:", err)
		return "", code.CodeServerBusy
	}
	return createdSession.ID, code.CodeSuccess
}

func StreamMessageToExistingSession(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {
	// 确保 writer 支持 Flush
	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Println("StreamMessageToExistingSession: streaming unsupported")
		return code.CodeServerBusy
	}

	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey":   "your-api-key", // TODO: 从配置中获取
		"username": userName,       // 用于 RAG 模型获取用户文档
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
	if err != nil {
		log.Println("StreamMessageToExistingSession GetOrCreateAIHelper error:", err)
		return code.AIModelFail
	}

	cb := func(msg string) {
		payload, err := json.Marshal(map[string]string{"delta": msg})
		if err != nil {
			log.Println("[SSE] Marshal error:", err)
			return
		}
		log.Printf("[SSE] Sending chunk len=%d", len(msg))
		_, err = writer.Write([]byte("data: " + string(payload) + "\n\n"))
		if err != nil {
			log.Println("[SSE] Write error:", err)
			return
		}
		flusher.Flush()
	}
	_, err_ := helper.StreamResponse(userName, ctx, cb, userQuestion)
	if err_ != nil {
		log.Println("StreamMessageToExistingSession StreamResponse error:", err_)
		return code.AIModelFail
	}

	_, err = writer.Write([]byte("data: [DONE]\n\n"))
	if err != nil {
		log.Println("StreamMessageToExistingSession write DONE error:", err)
		return code.AIModelFail
	}
	flusher.Flush()

	return code.CodeSuccess
}

func CreateStreamSessionAndSendMessage(userName string, userQuestion string, modelType string, writer http.ResponseWriter) (string, code.Code) {

	sessionID, code_ := CreateStreamSessionOnly(userName, userQuestion)
	if code_ != code.CodeSuccess {
		return "", code_
	}

	code_ = StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
	if code_ != code.CodeSuccess {

		return sessionID, code_
	}

	return sessionID, code.CodeSuccess
}

func ChatSend(userName string, sessionID string, userQuestion string, modelType string) (string, code.Code) {
	//1：获取AIHelper
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"username": userName, // 用于 RAG 模型获取用户文档（若当前用户选择了RAG模型，该字段将会被用到）
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
	if err != nil {
		log.Println("ChatSend GetOrCreateAIHelper error:", err)
		return "", code.AIModelFail
	}

	//2：生成AI回复
	aiResponse, err_ := helper.GenerateResponse(userName, ctx, userQuestion)
	if err_ != nil {
		log.Println("ChatSend GenerateResponse error:", err_)
		return "", code.AIModelFail
	}

	return aiResponse.Content, code.CodeSuccess
}

func GetChatHistory(userName string, sessionID string) ([]model.History, code.Code) {
	manager := aihelper.GetGlobalManager()
	if helper, exists := manager.GetAIHelper(userName, sessionID); exists {
		messages := helper.GetMessages()
		history := make([]model.History, 0, len(messages))
		for _, msg := range messages {
			history = append(history, model.History{IsUser: msg.IsUser, Content: msg.Content})
		}
		return history, code.CodeSuccess
	}

	msgs, err := messagedao.GetMessagesBySessionID(sessionID)
	if err != nil {
		log.Println("GetChatHistory GetMessagesBySessionID error:", err)
		return nil, code.CodeServerBusy
	}
	history := make([]model.History, 0, len(msgs))
	for _, msg := range msgs {
		if msg.UserName != userName {
			continue
		}
		history = append(history, model.History{IsUser: msg.IsUser, Content: msg.Content})
	}
	return history, code.CodeSuccess
}
func ChatStreamSend(userName string, sessionID string, userQuestion string, modelType string, writer http.ResponseWriter) code.Code {

	return StreamMessageToExistingSession(userName, sessionID, userQuestion, modelType, writer)
}

func RenameSession(userName, sessionID, title string) (*model.SessionInfo, code.Code) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, code.CodeInvalidParams
	}
	if utf8.RuneCountInString(title) > 60 {
		title = truncateRunes(title, 60)
	}
	if err := sessiondao.UpdateSessionTitle(userName, sessionID, title); err != nil {
		log.Println("RenameSession UpdateSessionTitle error:", err)
		return nil, code.CodeServerBusy
	}
	return &model.SessionInfo{SessionID: sessionID, Title: title}, code.CodeSuccess
}

func DeleteSession(userName, sessionID string) code.Code {
	if err := sessiondao.DeleteSession(userName, sessionID); err != nil {
		log.Println("DeleteSession error:", err)
		return code.CodeServerBusy
	}
	aihelper.GetGlobalManager().RemoveAIHelper(userName, sessionID)
	return code.CodeSuccess
}

func summarizeSessionTitle(question string) string {
	text := strings.Join(strings.Fields(question), " ")
	text = strings.Trim(text, " \t\r\n#*_`>~-,.!?;:()[]{}\"'")
	if text == "" {
		return "New Chat"
	}
	return truncateRunes(text, 24)
}
func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
