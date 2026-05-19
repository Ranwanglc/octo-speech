package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrompts_EmptyPath(t *testing.T) {
	ResetPromptsToDefaults()
	LoadPrompts("", nil)
	if activePrompts.System != systemPromptTemplate {
		t.Error("expected defaults with empty path")
	}
}

func TestLoadPrompts_NonExistentFile(t *testing.T) {
	ResetPromptsToDefaults()
	LoadPrompts("/nonexistent/path.yaml", nil)
	if activePrompts.System != systemPromptTemplate {
		t.Error("expected defaults with non-existent file")
	}
}

func TestLoadPrompts_ValidFile(t *testing.T) {
	ResetPromptsToDefaults()

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "prompts.yaml")

	content := `task_transcribe: "custom transcribe task"
task_edit: "custom edit task"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	LoadPrompts(yamlPath, nil)

	if activePrompts.TaskTranscribe != "custom transcribe task" {
		t.Errorf("expected custom task_transcribe, got %q", activePrompts.TaskTranscribe)
	}
	if activePrompts.TaskEdit != "custom edit task" {
		t.Errorf("expected custom task_edit, got %q", activePrompts.TaskEdit)
	}
	if activePrompts.System != systemPromptTemplate {
		t.Error("expected default system prompt (not overridden)")
	}
}

func TestLoadPrompts_InvalidYAML(t *testing.T) {
	ResetPromptsToDefaults()

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte(":::invalid:::"), 0644); err != nil {
		t.Fatal(err)
	}

	LoadPrompts(yamlPath, nil)

	if activePrompts.System != systemPromptTemplate {
		t.Error("expected defaults after invalid YAML")
	}
}

func TestLoadPrompts_TemplatePlaceholderValidation(t *testing.T) {
	ResetPromptsToDefaults()

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "prompts.yaml")

	content := `vocabulary_reference: "no placeholder here"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	LoadPrompts(yamlPath, nil)

	if activePrompts.VocabularyReference != vocabularyReferenceTemplate {
		t.Error("expected default vocabulary_reference when placeholder missing")
	}
}

func TestLoadPrompts_ValidTemplatePlaceholder(t *testing.T) {
	ResetPromptsToDefaults()

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "prompts.yaml")

	content := `vocabulary_reference: "custom vocab: %s"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	LoadPrompts(yamlPath, nil)

	if activePrompts.VocabularyReference != "custom vocab: %s" {
		t.Errorf("expected custom vocabulary_reference, got %q", activePrompts.VocabularyReference)
	}
}
