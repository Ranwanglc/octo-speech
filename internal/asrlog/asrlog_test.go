package asrlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogger_WriteEntry(t *testing.T) {
	dir := t.TempDir()

	logger := NewLogger(dir, 10, "test-pod", nil)
	if logger == nil {
		t.Fatal("expected logger to be created")
	}
	defer logger.Close()

	entry := ASREntry{
		RequestID: "test_123_abc",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Engine:    "gemini",
		AppID:     "app1",
		Input: ASRInput{
			MimeType:  "audio/wav",
			AudioSize: 100,
		},
		ResultText: "hello",
		AudioData:  []byte("fake audio data"),
	}

	logger.Enqueue(entry)
	logger.Close()

	dateDir := time.Now().UTC().Format("2006-01-02")
	jsonPath := filepath.Join(dir, dateDir, "gemini", "test_123_abc.json")
	audioPath := filepath.Join(dir, dateDir, "gemini", "test_123_abc.wav")

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("expected JSON file to exist")
	}
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Error("expected audio file to exist")
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}

	var saved ASREntry
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}

	if saved.AppID != "app1" {
		t.Errorf("expected app_id 'app1', got %q", saved.AppID)
	}
	if saved.PodID != "test-pod" {
		t.Errorf("expected pod_id 'test-pod', got %q", saved.PodID)
	}
	if saved.Input.AudioFile != "test_123_abc.wav" {
		t.Errorf("expected audio file ref, got %q", saved.Input.AudioFile)
	}
}

func TestLogger_GenerateRequestID(t *testing.T) {
	dir := t.TempDir()
	logger := NewLogger(dir, 10, "pod1", nil)
	if logger == nil {
		t.Fatal("expected logger")
	}
	defer logger.Close()

	id := logger.GenerateRequestID()
	if !strings.HasPrefix(id, "pod1_") {
		t.Errorf("expected prefix 'pod1_', got %q", id)
	}

	parts := strings.Split(id, "_")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in request ID, got %d: %q", len(parts), id)
	}
}

func TestLogger_NilOnBadDir(t *testing.T) {
	logger := NewLogger("/nonexistent/deeply/nested/path/that/cannot/exist", 10, "pod", nil)
	if logger != nil {
		logger.Close()
		t.Error("expected nil logger for non-writable directory")
	}
}

func TestCleaner_RemoveOld(t *testing.T) {
	dir := t.TempDir()

	oldDate := time.Now().UTC().AddDate(0, 0, -10).Format("2006-01-02")
	newDate := time.Now().UTC().Format("2006-01-02")

	os.MkdirAll(filepath.Join(dir, oldDate), 0755)
	os.MkdirAll(filepath.Join(dir, newDate), 0755)

	cleaner := NewCleaner(dir, 7, nil)
	cleaner.clean()

	if _, err := os.Stat(filepath.Join(dir, oldDate)); !os.IsNotExist(err) {
		t.Error("expected old dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, newDate)); os.IsNotExist(err) {
		t.Error("expected new dir to remain")
	}
}

func TestCleaner_StartAndClose(t *testing.T) {
	dir := t.TempDir()
	cleaner := NewCleaner(dir, 7, nil)
	cleaner.Start()
	time.Sleep(10 * time.Millisecond)
	cleaner.Close()
}

func TestMimeTypeToFormat(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"audio/wav", "wav"},
		{"audio/mpeg", "mp3"},
		{"audio/ogg", "ogg"},
		{"audio/webm", "webm"},
		{"audio/mp4", "m4a"},
		{"audio/flac", "flac"},
		{"unknown", "wav"},
	}

	for _, tt := range tests {
		got := mimeTypeToFormat(tt.mime)
		if got != tt.want {
			t.Errorf("mimeTypeToFormat(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}
