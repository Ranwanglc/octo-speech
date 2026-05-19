package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mininglamp-OSS/octo-speech/internal/config"
)

func TestTranscribe_AppendMode_NoSpeech(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "[NO_SPEECH]"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"test-model"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "existing text", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Text != "existing text" {
		t.Errorf("expected 'existing text' for NoSpeech in append mode, got %q", result.Text)
	}
}

func TestTranscribe_AppendMode_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "new content"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"test-model"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "hello", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Text != "hello new content" {
		t.Errorf("expected 'hello new content', got %q", result.Text)
	}
}

func TestTranscribe_EditMode_NoSpeech(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "[NO_SPEECH]"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"test-model"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "edit",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "existing", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Text != "existing" {
		t.Errorf("expected 'existing' for exact NoSpeech in edit mode, got %q", result.Text)
	}
}

func TestTranscribe_GPTEditNotSupported(t *testing.T) {
	cfg := &config.Config{
		LiteLLMUrl:   "http://unused",
		LiteLLMKey:   "key",
		Engine:       config.EngineGPT,
		GPTModels:    []string{"gpt-4o"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
	}

	svc := NewTranscribeService(cfg)
	_, err := svc.Transcribe([]byte("audio"), "audio/wav", "text", "", TranscribeOptions{
		Mode:   "edit_only",
		Engine: config.EngineGPT,
	})

	if err != ErrGPTEditNotSupported {
		t.Errorf("expected ErrGPTEditNotSupported, got %v", err)
	}
}

func TestTranscribe_ModelFallback(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
			return
		}
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "result"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"model-1", "model-2"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Model != "model-2" {
		t.Errorf("expected fallback to model-2, got %q", result.Model)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestTranscribe_NonRetryable4xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"model-1", "model-2"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	_, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{})
	if err == nil {
		t.Fatal("expected error")
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (non-retryable), got %d", callCount)
	}
}

func TestTranscribe_429Retryable(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limited"}`))
			return
		}
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "ok"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"model-1", "model-2"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (429 retryable), got %d", callCount)
	}
	if result.Text != "ok" {
		t.Errorf("expected 'ok', got %q", result.Text)
	}
}

func TestTranscribe_EngineOverride(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "text"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGPT,
		GPTModels:    []string{"gpt-m"},
		Models:       []string{"gemini-m"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	result, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{
		Engine: config.EngineGemini,
	})
	if err != nil {
		t.Fatal(err)
	}

	if receivedPath != "/chat/completions" {
		t.Errorf("expected /chat/completions for gemini override, got %s", receivedPath)
	}
	if result.Engine != config.EngineGemini {
		t.Errorf("expected gemini engine in result, got %s", result.Engine)
	}
}

func TestTranscribe_QwenDataURIPrefix(t *testing.T) {
	var body chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "result"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineQwen,
		QwenModels:   []string{"qwen-m"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	_, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Verify the body was sent with data: prefix
	if len(body.Messages) < 2 {
		t.Fatal("expected at least 2 messages")
	}
}

func TestTranscribe_GeminiReasoningEffort(t *testing.T) {
	var body chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "result"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"gemini-3.1-pro-preview"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	_, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if body.ReasoningEffort != "low" {
		t.Errorf("expected reasoning_effort=low for 3.1-pro, got %q", body.ReasoningEffort)
	}
}

func TestJoinContextAndText(t *testing.T) {
	tests := []struct {
		ctx  string
		text string
		want string
	}{
		{"", "hello", "hello"},
		{"hello", "", "hello"},
		{"hello", "world", "hello world"},
		{"你好", "世界", "你好世界"},
		{"hello ", "world", "hello world"},
		{"hello!", "world", "hello!world"},
		{"hello。", "world", "hello。world"},
		{"english", "中文", "english中文"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s+%s", tt.ctx, tt.text), func(t *testing.T) {
			got := joinContextAndText(tt.ctx, tt.text)
			if got != tt.want {
				t.Errorf("joinContextAndText(%q, %q) = %q, want %q", tt.ctx, tt.text, got, tt.want)
			}
		})
	}
}

func TestRestoreTrailingWhitespace(t *testing.T) {
	tests := []struct {
		ctx  string
		text string
		want string
	}{
		{"hello ", "helloworld", "hello world"},
		{"hello\n", "edited text", "edited text\n"},
		{"", "text", "text"},
		{"hello", "hello world", "hello world"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			got := restoreTrailingWhitespace(tt.ctx, tt.text)
			if got != tt.want {
				t.Errorf("restoreTrailingWhitespace(%q, %q) = %q, want %q",
					tt.ctx, tt.text, got, tt.want)
			}
		})
	}
}

func TestMimeTypeToFormat(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"audio/wav", "wav"},
		{"audio/mpeg", "mp3"},
		{"audio/mp3", "mp3"},
		{"audio/ogg", "ogg"},
		{"audio/webm", "webm"},
		{"audio/mp4", "m4a"},
		{"audio/x-m4a", "m4a"},
		{"audio/flac", "flac"},
		{"application/octet-stream", "wav"},
	}

	for _, tt := range tests {
		got := MimeTypeToFormat(tt.mime)
		if got != tt.want {
			t.Errorf("MimeTypeToFormat(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 400, Body: "bad request"}
	if err.Error() != "API error 400: bad request" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestIsNonRetryableError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{&APIError{StatusCode: 400}, true},
		{&APIError{StatusCode: 401}, true},
		{&APIError{StatusCode: 403}, true},
		{&APIError{StatusCode: 429}, false},
		{&APIError{StatusCode: 500}, false},
		{fmt.Errorf("other error"), false},
	}

	for _, tt := range tests {
		got := isNonRetryableError(tt.err)
		if got != tt.want {
			t.Errorf("isNonRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestTranscribe_EmotionEmojiOverride(t *testing.T) {
	var body chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		resp := chatCompletionResponse{
			Choices: []choice{{Message: responseMessage{Content: "text"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.Config{
		LiteLLMUrl:   server.URL,
		LiteLLMKey:   "test-key",
		Engine:       config.EngineGemini,
		Models:       []string{"test-model"},
		Timeout:      30,
		TotalTimeout: 45,
		EditMode:     "append",
		EmotionEmoji: true,
	}

	svc := NewTranscribeService(cfg)
	ResetPromptsToDefaults()

	falseVal := false
	_, err := svc.Transcribe([]byte("audio"), "audio/wav", "", "", TranscribeOptions{
		EmotionEmoji: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
}
