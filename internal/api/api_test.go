package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Mininglamp-OSS/octo-speech/internal/config"
	"github.com/Mininglamp-OSS/octo-speech/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthRouter(appStore *store.AppStore) *gin.Engine {
	r := gin.New()
	r.Use(AuthMiddleware(appStore))
	r.GET("/test", func(c *gin.Context) {
		appID, _ := c.Get("app_id")
		c.JSON(200, gin.H{"app_id": appID})
	})
	return r
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	s := store.NewAppStore(nil, 60)
	r := setupAuthRouter(s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["msg"] != "missing authorization header" {
		t.Errorf("unexpected msg: %v", resp["msg"])
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	s := store.NewAppStore(nil, 60)
	r := setupAuthRouter(s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestVocabularyHandler_PutValidation(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})

	h := NewVocabularyHandler(nil)
	r.PUT("/v1/speech/vocabularies", h.Put)

	tests := []struct {
		name string
		body string
		want int
	}{
		{
			"missing subject_id",
			`{"scope_type":"global","scope_id":"default","content":"test"}`,
			400,
		},
		{
			"missing content",
			`{"subject_id":"user1","scope_type":"global","scope_id":"default"}`,
			400,
		},
		{
			"invalid scope_type",
			`{"subject_id":"user1","scope_type":"invalid","scope_id":"default","content":"test"}`,
			400,
		},
		{
			"empty content",
			`{"subject_id":"user1","scope_type":"global","scope_id":"default","content":""}`,
			400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/v1/speech/vocabularies",
				strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("expected %d, got %d: %s", tt.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestVocabularyHandler_GetValidation(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})

	h := NewVocabularyHandler(nil)
	r.GET("/v1/speech/vocabularies", h.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/speech/vocabularies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 without subject_id, got %d", w.Code)
	}
}

func TestVocabularyHandler_GetInvalidScope(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})

	h := NewVocabularyHandler(nil)
	r.GET("/v1/speech/vocabularies", h.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/speech/vocabularies?subject_id=u1&scope_type=bad", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid scope_type, got %d", w.Code)
	}
}

func TestVocabularyHandler_DeleteValidation(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})

	h := NewVocabularyHandler(nil)
	r.DELETE("/v1/speech/vocabularies", h.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/v1/speech/vocabularies?subject_id=u1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 without all params, got %d", w.Code)
	}
}

func TestConfigHandler(t *testing.T) {
	cfg := &config.Config{
		MaxDuration:        60,
		MaxFileSize:        3145728,
		Engine:             config.EngineGemini,
		EditMode:           "edit",
		LocalEnabled:       true,
		LocalTimeoutMs:     10000,
		LocalProbeURL:      "http://localhost:8787/",
		LocalTranscribeURL: "http://localhost:8787/v1/voice/transcribe",
	}

	localStore := store.NewLocalConfigStore(nil, cfg)
	h := NewConfigHandler(cfg, localStore)

	r := gin.New()
	r.GET("/v1/speech/config", h.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/speech/config", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["engine"] != "gm" {
		t.Errorf("expected engine 'gm', got %v", resp["engine"])
	}
	if resp["edit_mode"] != "edit" {
		t.Errorf("expected edit_mode 'edit', got %v", resp["edit_mode"])
	}
	if resp["max_duration"] != float64(60) {
		t.Errorf("expected max_duration 60, got %v", resp["max_duration"])
	}
	if resp["local_enabled"] != true {
		t.Errorf("expected local_enabled true, got %v", resp["local_enabled"])
	}
}

func TestTranscribeHandler_MissingAudio(t *testing.T) {
	cfg := &config.Config{
		MaxFileSize:  3145728,
		EmotionEmoji: true,
	}

	h := NewTranscribeHandler(nil, cfg, nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})
	r.POST("/v1/speech/transcribe", h.Handle)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/speech/transcribe", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTranscribeHandler_InvalidMode(t *testing.T) {
	cfg := &config.Config{
		MaxFileSize:            3145728,
		EmotionEmoji:           true,
		MaxContextTextLength:   5000,
		MaxChatContextLength:   20000,
		MaxVoiceContextLength:  10000,
		MaxMemberContextLength: 5000,
	}

	h := NewTranscribeHandler(nil, cfg, nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})
	r.POST("/v1/speech/transcribe", h.Handle)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("audio", "test.wav")
	part.Write([]byte("fake audio"))
	writer.WriteField("mode", "invalid_mode")
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/speech/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid mode, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTranscribeHandler_InvalidEngine(t *testing.T) {
	cfg := &config.Config{
		MaxFileSize:            3145728,
		EmotionEmoji:           true,
		MaxContextTextLength:   5000,
		MaxChatContextLength:   20000,
		MaxVoiceContextLength:  10000,
		MaxMemberContextLength: 5000,
	}

	h := NewTranscribeHandler(nil, cfg, nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("app_id", "test-app")
		c.Next()
	})
	r.POST("/v1/speech/transcribe", h.Handle)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("audio", "test.wav")
	part.Write([]byte("fake audio"))
	writer.WriteField("engine", "invalid_engine")
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/speech/transcribe", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid engine, got %d: %s", w.Code, w.Body.String())
	}
}
