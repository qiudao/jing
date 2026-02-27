package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	q := NewQTable()
	q.Set("s1", 0, 0.5)
	q.Set("s1", 4, 0.9)
	q.Set("s2", 2, -0.3)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	err := SaveQTable(q, path)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadQTable(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Get("s1", 0) != 0.5 {
		t.Errorf("expected 0.5, got %f", loaded.Get("s1", 0))
	}
	if loaded.Get("s1", 4) != 0.9 {
		t.Errorf("expected 0.9, got %f", loaded.Get("s1", 4))
	}
	if loaded.Get("s2", 2) != -0.3 {
		t.Errorf("expected -0.3, got %f", loaded.Get("s2", 2))
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := LoadQTable("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	q := NewQTable()
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "model.json")
	err := SaveQTable(q, path)
	if err != nil {
		t.Fatalf("save should create parent dirs: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}
