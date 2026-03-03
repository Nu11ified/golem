package server_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nu11ified/golem/src/server"
)

func TestSaveAndLoadWorkspace(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	ws := map[string]interface{}{
		"pages":       map[string]interface{}{},
		"rootPageIds": []interface{}{},
	}
	if err := s.SaveWorkspace(ws); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil workspace")
	}
}

func TestLoadWorkspaceNotExist(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	loaded, err := s.LoadWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Error("expected nil for non-existent workspace")
	}
}

func TestSaveAndLoadBlocks(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	blocks := []interface{}{
		map[string]interface{}{"id": "b1", "type": "text", "content": "Hello"},
	}
	if err := s.SavePageBlocks("page-1", blocks); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadPageBlocks("page-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 block, got %d", len(loaded))
	}
}

func TestLoadBlocksNotExist(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	loaded, err := s.LoadPageBlocks("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Error("expected nil for non-existent blocks")
	}
}

func TestDeletePageData(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	if err := s.SavePageBlocks("page-1", []interface{}{}); err != nil {
		t.Fatal(err)
	}

	if err := s.DeletePageData("page-1"); err != nil {
		t.Fatal(err)
	}

	blocksFile := filepath.Join(dir, "blocks", "page-1.json")
	if _, err := os.Stat(blocksFile); !os.IsNotExist(err) {
		t.Error("blocks file should be deleted")
	}
}

func TestDeletePageDataNotExist(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	// Should not error when file doesn't exist
	if err := s.DeletePageData("nonexistent"); err != nil {
		t.Fatal(err)
	}
}

func TestSaveWorkspaceOverwrite(t *testing.T) {
	dir := t.TempDir()
	s := server.NewStorage(dir)

	ws1 := map[string]interface{}{"version": float64(1)}
	ws2 := map[string]interface{}{"version": float64(2)}

	if err := s.SaveWorkspace(ws1); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveWorkspace(ws2); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.LoadWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	m := loaded.(map[string]interface{})
	if m["version"] != float64(2) {
		t.Errorf("expected version 2, got %v", m["version"])
	}
}
