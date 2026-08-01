package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// SentVideo captures data from a sendVideo API call.
type SentVideo struct {
	ChatID  string
	ReplyTo string
	Video   []byte
}

// SentMessage captures data from a sendMessage API call.
type SentMessage struct {
	ChatID  string
	Text    string
	ReplyTo string
}

// MockTelegramServer is a fake Telegram Bot API server for e2e tests.
type MockTelegramServer struct {
	Server *httptest.Server

	mu              sync.Mutex
	SentVideos      []SentVideo
	SentMessages    []SentMessage
	ChatActionsSent []string
}

// NewMockTelegramServer creates and starts a mock Telegram API server.
func NewMockTelegramServer(t *testing.T) *MockTelegramServer {
	t.Helper()
	m := &MockTelegramServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handleRequest)

	m.Server = httptest.NewServer(mux)
	t.Cleanup(m.Server.Close)
	return m
}

func (m *MockTelegramServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract method from path: /bot<token>/<method>
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		m.writeOK(w, json.RawMessage(`{}`))
		return
	}
	method := parts[len(parts)-1]

	switch method {
	case "getMe":
		m.handleGetMe(w)
	case "sendVideo":
		m.handleSendVideo(w, r)
	case "sendMessage":
		m.handleSendMessage(w, r)
	case "sendChatAction":
		m.handleSendChatAction(w, r)
	case "setMyCommands", "deleteMessage", "getUpdates", "deleteMyCommands":
		m.writeOK(w, json.RawMessage(`true`))
	default:
		m.writeOK(w, json.RawMessage(`true`))
	}
}

func (m *MockTelegramServer) handleGetMe(w http.ResponseWriter) {
	result := json.RawMessage(`{
		"id": 123456789,
		"is_bot": true,
		"first_name": "TestBot",
		"username": "test_bot"
	}`)
	m.writeOK(w, result)
}

// getJSONBody parses the request body as a JSON map. Safe to call multiple times.
func getJSONBody(r *http.Request) map[string]interface{} {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return data
}

func jsonMapString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return fmt.Sprintf("%.0f", val)
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

func (m *MockTelegramServer) handleSendVideo(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20) // 50MB max

	chatID := r.FormValue("chat_id")
	replyTo := r.FormValue("reply_to_message_id")

	var videoBytes []byte
	// First try file fields
	if r.MultipartForm != nil {
		for _, files := range r.MultipartForm.File {
			if len(files) > 0 {
				file, err := files[0].Open()
				if err == nil {
					videoBytes, _ = io.ReadAll(file)
					file.Close()
					break
				}
			}
		}
	}
	// Telebot v4 may send file content as a form value instead of a file field
	if len(videoBytes) == 0 {
		if v := r.FormValue("video"); v != "" {
			videoBytes = []byte(v)
		}
	}

	m.mu.Lock()
	m.SentVideos = append(m.SentVideos, SentVideo{
		ChatID:  chatID,
		ReplyTo: replyTo,
		Video:   videoBytes,
	})
	m.mu.Unlock()

	// Use a safe default for chat_id in response
	responseChatID := chatID
	if responseChatID == "" {
		responseChatID = "0"
	}

	result := json.RawMessage(`{
		"message_id": 42,
		"chat": {"id": ` + responseChatID + `},
		"video": {"file_id": "fake_file_id"}
	}`)
	m.writeOK(w, result)
}

func (m *MockTelegramServer) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var chatID, text, replyTo string

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		data := getJSONBody(r)
		chatID = jsonMapString(data, "chat_id")
		text = jsonMapString(data, "text")

		// reply_parameters is an object with message_id
		if rp, ok := data["reply_parameters"]; ok {
			if rpMap, ok := rp.(map[string]interface{}); ok {
				replyTo = jsonMapString(rpMap, "message_id")
			}
		}
	} else {
		r.ParseForm()
		chatID = r.FormValue("chat_id")
		text = r.FormValue("text")
		replyTo = r.FormValue("reply_to_message_id")
	}

	m.mu.Lock()
	m.SentMessages = append(m.SentMessages, SentMessage{
		ChatID:  chatID,
		Text:    text,
		ReplyTo: replyTo,
	})
	m.mu.Unlock()

	responseChatID := chatID
	if responseChatID == "" {
		responseChatID = "0"
	}

	result := json.RawMessage(`{
		"message_id": 43,
		"chat": {"id": ` + responseChatID + `},
		"text": ` + jsonString(text) + `
	}`)
	m.writeOK(w, result)
}

func (m *MockTelegramServer) handleSendChatAction(w http.ResponseWriter, r *http.Request) {
	var action string
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		data := getJSONBody(r)
		action = jsonMapString(data, "action")
	} else {
		r.ParseForm()
		action = r.FormValue("action")
	}

	m.mu.Lock()
	m.ChatActionsSent = append(m.ChatActionsSent, action)
	m.mu.Unlock()

	m.writeOK(w, json.RawMessage(`true`))
}

func (m *MockTelegramServer) writeOK(w http.ResponseWriter, result json.RawMessage) {
	resp := map[string]interface{}{
		"ok":     true,
		"result": result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
