package asrlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ASRInput struct {
	Mode            string `json:"mode"`
	MimeType        string `json:"mime_type"`
	AudioSize       int    `json:"audio_size"`
	AudioFile       string `json:"audio_file"`
	ContextText     string `json:"context_text"`
	ChatContext     string `json:"chat_context"`
	PersonalContext string `json:"personal_context"`
	MemberContext   string `json:"member_context"`
	Model           string `json:"model"`
	Language        string `json:"language"`
	ChannelType     string `json:"channel_type,omitempty"`
}

type ASRPrompt struct {
	Type        string      `json:"type"`
	Text        string      `json:"text"`
	RequestBody interface{} `json:"request_body"`
}

type ASREntry struct {
	RequestID      string     `json:"request_id"`
	Timestamp      string     `json:"timestamp"`
	Source         string     `json:"source"`
	AppID          string     `json:"app_id"`
	Engine         string     `json:"engine"`
	ModelRequested string     `json:"model_requested"`
	ModelUsed      string     `json:"model_used"`
	Input          ASRInput   `json:"input"`
	Prompt         *ASRPrompt `json:"prompt,omitempty"`
	RawResultText  string     `json:"raw_result_text"`
	ResultText     string     `json:"result_text"`
	ResultLength   int        `json:"result_length"`
	IsNoSpeech     bool       `json:"is_no_speech"`
	Error          string     `json:"error"`
	DurationMs     int64      `json:"duration_ms"`
	PodID          string     `json:"pod_id"`

	AudioData []byte `json:"-"`
}

type Logger struct {
	baseDir string
	podID   string
	ch      chan ASREntry
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	logger  *zap.Logger
}

func NewLogger(baseDir string, bufSize int, podID string, logger *zap.Logger) *Logger {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		if logger != nil {
			logger.Error("ASR_LOG_DIR is not writable, ASR logging disabled",
				zap.String("dir", baseDir), zap.Error(err))
		}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	l := &Logger{
		baseDir: baseDir,
		podID:   podID,
		ch:      make(chan ASREntry, bufSize),
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
	}
	l.wg.Add(1)
	go l.worker()
	return l
}

func (l *Logger) Close() {
	l.cancel()
	l.wg.Wait()
}

func (l *Logger) Enqueue(entry ASREntry) {
	if l.ctx.Err() != nil {
		return
	}
	entry.PodID = l.podID
	select {
	case l.ch <- entry:
	case <-l.ctx.Done():
	default:
		if l.logger != nil {
			l.logger.Warn("asr log buffer full, dropping entry",
				zap.String("request_id", entry.RequestID))
		}
	}
}

func (l *Logger) GenerateRequestID() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%s_%d_%s", l.podID,
		time.Now().UTC().UnixMilli(),
		hex.EncodeToString(b))
}

func (l *Logger) worker() {
	defer l.wg.Done()

	for {
		select {
		case entry := <-l.ch:
			l.writeEntry(entry)
		case <-l.ctx.Done():
			goto drain
		}
	}

drain:
	for {
		select {
		case entry := <-l.ch:
			l.writeEntry(entry)
		default:
			return
		}
	}
}

func mimeTypeToFormat(mimeType string) string {
	switch {
	case strings.Contains(mimeType, "wav"):
		return "wav"
	case strings.Contains(mimeType, "mp3"), strings.Contains(mimeType, "mpeg"):
		return "mp3"
	case strings.Contains(mimeType, "ogg"):
		return "ogg"
	case strings.Contains(mimeType, "webm"):
		return "webm"
	case strings.Contains(mimeType, "mp4"), strings.Contains(mimeType, "m4a"):
		return "m4a"
	case strings.Contains(mimeType, "flac"):
		return "flac"
	default:
		return "wav"
	}
}

func (l *Logger) writeEntry(entry ASREntry) {
	now, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
	if now.IsZero() {
		now = time.Now().UTC()
	}

	dateDir := now.UTC().Format("2006-01-02")
	dir := filepath.Join(l.baseDir, dateDir, entry.Engine)

	if err := os.MkdirAll(dir, 0755); err != nil {
		if l.logger != nil {
			l.logger.Error("mkdir failed", zap.Error(err), zap.String("dir", dir))
		}
		return
	}

	if len(entry.AudioData) > 0 {
		ext := mimeTypeToFormat(entry.Input.MimeType)
		audioFileName := entry.RequestID + "." + ext
		audioPath := filepath.Join(dir, audioFileName)
		if err := os.WriteFile(audioPath, entry.AudioData, 0644); err != nil {
			if l.logger != nil {
				l.logger.Error("write audio file failed", zap.Error(err),
					zap.String("path", audioPath))
			}
		}
		entry.Input.AudioFile = audioFileName
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		if l.logger != nil {
			l.logger.Error("marshal metadata failed", zap.Error(err))
		}
		return
	}
	jsonPath := filepath.Join(dir, entry.RequestID+".json")
	if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
		if l.logger != nil {
			l.logger.Error("write metadata failed", zap.Error(err))
		}
	}
}
